FROM alpine:latest AS builder

# 设置国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories

# 更新系统并安装基本的构建依赖
RUN apk update && \
    apk add --no-cache \
    nodejs \
    npm \
    wget \
    tar \
    build-base \
    git \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# 设置npm镜像源
RUN npm config set registry https://registry.npmmirror.com

ARG GO_VERSION=1.25.5
ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH

RUN wget https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -O go.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz

# 设置go镜像源
ENV GOPROXY=https://goproxy.cn,direct

# fix sqllite问题
RUN go env -w CGO_CFLAGS='-O2 -g -D_LARGEFILE64_SOURCE'

WORKDIR /app

# 前端构建
COPY ./web/package.json ./web/package-lock.json ./web/
RUN cd ./web && npm ci
COPY ./web/ ./web/
RUN cd ./web && npm run build

# Go 依赖缓存层
COPY ./go.mod ./go.sum ./
RUN go mod download

# Go 源代码
COPY ./client ./client/
COPY ./cmd ./cmd/
COPY ./internal ./internal/
COPY ./pkg ./pkg

# 构建服务端
RUN go build -o ./app ./cmd/main.go

# === 运行阶段 ===
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/web/dist ./web/dist
COPY --from=builder /app/app /app/app
COPY ./configs ./configs
COPY ./pkg/ip/GeoLite2-Country.mmdb ./pkg/ip/GeoLite2-Country.mmdb

RUN mkdir -p ./logs ./data ./tmp

EXPOSE 8080 8081 8082

ENV NEXUS_SERVER_HOST=0.0.0.0

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./app"]