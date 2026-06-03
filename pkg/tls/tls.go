package tls

import (
	"crypto/tls"
	"fmt"
	"net"
)

// ServerConfig 服务端 TLS 配置
type ServerConfig struct {
	CertFile string
	KeyFile  string
}

// LoadServerConfig 从证书和密钥文件加载服务端 TLS 配置。
func LoadServerConfig(cfg *ServerConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS key pair: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// WrapListener 将 net.Listener 包装为 TLS listener。
func WrapListener(ln net.Listener, config *tls.Config) net.Listener {
	return tls.NewListener(ln, config)
}
