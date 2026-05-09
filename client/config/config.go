package config

type Config struct {
	ServerAddr        string `json:"server_addr"` // 服务器地址
	AppId             int64  `json:"app_id"`      // 应用id
	AppSecret         string `json:"app_secret"`  // 应用密钥
	AutoUpgrade       bool   `json:"auto_upgrade"`
	ReconnectInterval uint32 `json:"reconnect_interval"`
	HeartbeatInterval uint32 `json:"heartbeat_interval"`
	ConnectTimeout    uint32 `json:"connect_timeout"`
}
