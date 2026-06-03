package config

import "fmt"

type Config struct {
	ServerAddr        string `json:"server_addr"` // 服务器地址
	AppId             int64  `json:"app_id"`      // 应用id
	AppSecret         string `json:"app_secret"`  // 应用密钥
	ReconnectInterval uint32 `json:"reconnect_interval"`
	HeartbeatInterval uint32 `json:"heartbeat_interval"`
	ConnectTimeout    uint32 `json:"connect_timeout"`

	TLSEnabled    bool   `json:"tls_enabled"`     // 是否启用 TLS
	TLSCACert     string `json:"tls_ca_cert"`     // CA 证书路径（自签名证书时使用）
	TLSServerName string `json:"tls_server_name"` // SNI 服务器名称
}

func (c *Config) Validate() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("server_addr is required")
	}
	if c.AppId == 0 {
		return fmt.Errorf("app_id is required")
	}
	if c.AppSecret == "" {
		return fmt.Errorf("app_secret is required")
	}
	if c.ReconnectInterval == 0 {
		c.ReconnectInterval = 10
	}
	if c.HeartbeatInterval == 0 {
		c.HeartbeatInterval = 30
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 10
	}
	return nil
}
