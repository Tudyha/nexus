package config

type Config struct {
	ServerAddr        string `json:"server_addr"` // 服务器地址
	AppId             int64  `json:"app_id"`      // 应用id
	AppSecret         string `json:"app_secret"`  // 应用密钥
	Version           int32  // 客户端版本
	ReconnectInterval uint32 // 重连间隔
	HeartbeatInterval uint32 // 心跳间隔
	ConnectTimeout    uint32 // 连接超时时间
}
