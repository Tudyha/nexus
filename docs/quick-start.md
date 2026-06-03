# Nexus 内网穿透 — 快速开始

## 概述

Nexus 是一款内网穿透工具，可以将内网服务通过公网服务器暴露到外网。
客户端部署在内网机器上，与服务端建立长连接，服务端将外部请求通过隧道转发到内网服务。

工作流程：`外网用户 -> 服务端(公网) -> 隧道 -> 客户端(内网) -> 本地服务`

---

## 一、服务端部署

### 方式 1：Docker Compose（推荐）

```bash
# 克隆项目
git clone <repo-url> && cd nexus

# 修改配置（可选）
vim configs/config.yaml

# 启动服务
docker compose up -d

# 查看日志
docker compose logs -f
```

### 方式 2：手动编译运行

```bash
# 编译
make build

# 运行
./bin/nexus-server

# 或直接
go run ./cmd/main.go
```

### 方式 3：源码直接运行

```bash
go run ./cmd/main.go
```

### 服务端配置

编辑 `configs/config.yaml`：

```yaml
server:
  http:
    port: 8080    # 管理面板 + API
  tcp:
    port: 8081    # 客户端接入端口
  v2ray:
    port: 8082    # V2Ray 代理端口（可选）

database:
  type: sqlite
  dns: data/nexus.db
```

启动后访问 `http://<服务器IP>:8080` 进入管理面板。

---

## 二、管理后台操作

### 1. 首次使用

1. 浏览器打开 `http://<服务器IP>:8080`
2. 注册管理员账号
3. 登录管理面板

### 2. 创建应用（App）

1. 进入「应用管理」页面
2. 点击「创建应用」
3. 填写应用名称
4. 创建后系统会生成 `app_id` 和 `app_secret`
5. **记录这两个值**，客户端连接时需要用到

### 3. 创建隧道

1. 进入「隧道管理」页面
2. 点击「创建隧道」
3. 填写以下信息：
   - **客户端**：选择目标客户端
   - **隧道类型**：TCP / UDP
   - **本地端口**：内网服务的端口
   - **远程端口**：公网服务器上暴露的端口

示例：将内网 192.168.1.10:22（SSH）暴露到公网服务器的 2222 端口：
   - 隧道类型：TCP
   - 本地地址：192.168.1.10:22
   - 远程端口：2222
   - 外网用户通过 `服务器IP:2222` 即可 SSH 到内网机器

---

## 三、客户端部署

### 下载客户端

从服务端管理面板下载对应平台的客户端二进制文件，或自行编译：

```bash
# 编译所有平台客户端
make build-client-all

# 编译当前平台
cd client && go build -o nexus-cli ./main.go
```

### 客户端配置

创建 `config.json`：

```json
{
  "server_addr": "your-server.com:8081",
  "app_id": 1,
  "app_secret": "your-app-secret",
  "reconnect_interval": 5,
  "heartbeat_interval": 30,
  "connect_timeout": 10
}
```

| 字段 | 说明 |
|------|------|
| `server_addr` | 服务端地址和端口（TCP 端口，非 HTTP 端口） |
| `app_id` | 应用 ID，在管理后台创建应用后获得 |
| `app_secret` | 应用密钥，创建应用后获得 |
| `reconnect_interval` | 断线重连间隔（秒） |
| `heartbeat_interval` | 心跳间隔（秒） |
| `connect_timeout` | 连接超时（秒） |

### 启动客户端

```bash
# 前台启动
./nexus-cli run -f config.json

# 后台守护进程模式
./nexus-cli run -d -f config.json

# 查看状态
./nexus-cli status

# 查看日志
./nexus-cli logs

# 查看实时日志
./nexus-cli logs -f

# 停止客户端
./nexus-cli stop

# 重启客户端
./nexus-cli restart -f config.json
```

---

## 四、完整示例

将内网的 Nginx Web 服务暴露到公网：

### 服务端操作

| 步骤 | 操作 |
|------|------|
| 1 | 部署服务端并启动 |
| 2 | 登录管理面板（http://服务器IP:8080） |
| 3 | 创建应用，记录 app_id 和 app_secret |
| 4 | 创建 TCP 隧道：本地端口 80 → 远程端口 8088 |

### 客户端操作

```bash
# 创建配置文件
cat > config.json <<EOF
{
  "server_addr": "your-server.com:8081",
  "app_id": 1,
  "app_secret": "xxx",
  "reconnect_interval": 5,
  "heartbeat_interval": 30,
  "connect_timeout": 10
}
EOF

# 启动
./nexus-cli run -d -f config.json
```

### 验证

外网用户访问 `http://your-server.com:8088` 即可看到内网 Nginx 页面。

---

## 五、常用命令速查

```bash
# 服务端
make build                              # 编译
make docker                              # 构建 Docker 镜像
go run ./cmd/main.go                     # 源码运行

# 客户端
nexus-cli run -f config.json             # 前台启动
nexus-cli run -d -f config.json          # 后台启动
nexus-cli status                         # 查看状态
nexus-cli logs [-f]                      # 查看日志
nexus-cli stop                           # 停止
nexus-cli restart -f config.json         # 重启

# 编译
make build-client-all                    # 交叉编译所有平台
```

---

## 六、常见问题

### 客户端连接不上服务端

1. 检查 `server_addr` 是否正确（使用 TCP 端口 8081，不是 HTTP 端口 8080）
2. 检查服务端防火墙是否放行了 TCP 端口
3. 检查 `app_id` 和 `app_secret` 是否正确

### 隧道无法访问

1. 确认客户端状态为在线（管理后台可见）
2. 检查远程端口是否被防火墙拦截
3. 检查内网目标服务是否正常启动
4. 尝试使用 `localhost` 而非 `127.0.0.1` 作为本地地址

### 客户端频繁断连

1. 检查网络稳定性
2. 适当增大 `heartbeat_interval` 和 `reconnect_interval`
3. 检查服务端 `keep_alive` 配置
