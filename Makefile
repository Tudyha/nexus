# === Go 基础 ===
.PHONY: build-server build-client build proto lint test vet clean

build-server:
	go build -o bin/nexus-server ./cmd/main.go

build-client:
	cd client && go build -o ../bin/nexus-cli ./main.go

build: build-server build-client

proto:
	cd pkg/proto && protoc --go_out=. --go_opt=paths=source_relative ./*.proto

lint:
	golangci-lint run ./...
	cd web && npx eslint --ext .ts,.vue src/

test:
	go test -v -race -count=1 ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

# === Docker ===
.PHONY: docker docker-push

docker:
	docker build -t nexus:latest .

docker-push: docker
	docker tag nexus:latest registry.cn-guangzhou.aliyuncs.com/knodio/nexus:latest
	docker push registry.cn-guangzhou.aliyuncs.com/knodio/nexus:latest

# === 客户端交叉编译 ===
.PHONY: build-client-all

build-client-all:
	cd client && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ../build/nexus-cli-darwin-amd64 ./main.go
	cd client && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ../build/nexus-cli-darwin-arm64 ./main.go
	cd client && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../build/nexus-cli-linux-amd64 ./main.go
	cd client && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ../build/nexus-cli-linux-arm64 ./main.go
	cd client && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ../build/nexus-cli-windows-amd64.exe ./main.go

# === Web 前端 ===
.PHONY: web-install web-dev web-build

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# === 全部 ===
.PHONY: all

all: vet test build web-build docker