package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

var cfg *Config

// Config 配置结构体
type Config struct {
	Server ServerConfig   `mapstructure:"server"`   // 服务配置
	DB     DatabaseConfig `mapstructure:"database"` // 数据库配置
}

// ServerConfig 服务配置
type ServerConfig struct {
	Env       string          `mapstructure:"env"`        // 环境: dev, prod
	Host      string          `mapstructure:"host"`       // 服务监听地址
	JWTSecret string          `mapstructure:"jwt_secret"` // JWT 签名密钥（留空则启动时生成随机密钥）
	TLS       TLSConfig       `mapstructure:"tls"`        // TLS 配置（可选）
	LogLevel  string          `mapstructure:"log_level"`  // 日志级别: debug, info, warn, error（默认 info）
	RateLimit RateLimitConfig `mapstructure:"rate_limit"` // API 速率限制配置
	HTTP      HttpConfig      `mapstructure:"http"`       // http服务配置
	TCP       TCPConfig       `mapstructure:"tcp"`        // tcp服务配置
	V2ray     V2rayConfig     `mapstructure:"v2ray"`      // v2ray服务配置
}

// RateLimitConfig API 速率限制配置
type RateLimitConfig struct {
	Enabled bool    `mapstructure:"enabled"` // 是否启用
	Rate    float64 `mapstructure:"rate"`    // 每秒令牌数
	Burst   int     `mapstructure:"burst"`   // 最大突发请求
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`   // 是否启用 TLS
	CertFile string `mapstructure:"cert_file"` // 证书文件路径
	KeyFile  string `mapstructure:"key_file"`  // 密钥文件路径
}

type HttpConfig struct {
	Port         int32 `mapstructure:"port"` // http监听端口
	ReadTimeout  int32 `mapstructure:"read_timeout"`
	WriteTimeout int32 `mapstructure:"write_timeout"`
}

// TCPConfig TCP服务配置
type TCPConfig struct {
	Port         int32 `mapstructure:"port"`          // tcp监听端口
	KeepAlive    int32 `mapstructure:"keep_alive"`    // tcp连接保持时间，单位秒
	ReadTimeout  int32 `mapstructure:"read_timeout"`  // tcp读取超时时间，单位秒
	WriteTimeout int32 `mapstructure:"write_timeout"` // tcp写入超时时间，单位秒
}

type V2rayConfig struct {
	Port    int32  `mapstructure:"port"`     // v2ray监听端口
	LogPath string `mapstructure:"log_path"` // v2ray日志路径
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string `mapstructure:"type"`              // 数据库类型: sqlite
	DNS             string `mapstructure:"dns"`               // 数据库连接字符串
	LogLevel        string `mapstructure:"log_level"`         // 日志级别
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`    // 最大空闲连接数
	MaxOpenConns    int    `mapstructure:"max_open_conns"`    // 最大打开连接数
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 连接最大生命周期，单位秒
}

// Init 初始化配置
func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 绑定环境变量
	viper.AutomaticEnv()
	viper.SetEnvPrefix("NEXUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 解析配置
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	return nil
}

// Validate 验证配置合法性
func (c *Config) Validate() error {
	var errs []error

	// HTTP 配置
	if c.Server.HTTP.Port <= 0 || c.Server.HTTP.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.http.port 必须在 1-65535 之间"))
	}
	if c.Server.HTTP.ReadTimeout < 1 {
		errs = append(errs, fmt.Errorf("server.http.read_timeout 不能小于 1 秒"))
	}
	if c.Server.HTTP.WriteTimeout < 1 {
		errs = append(errs, fmt.Errorf("server.http.write_timeout 不能小于 1 秒"))
	}

	// TCP 配置
	if c.Server.TCP.Port <= 0 || c.Server.TCP.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.tcp.port 必须在 1-65535 之间"))
	}
	if c.Server.TCP.KeepAlive < 10 {
		errs = append(errs, fmt.Errorf("server.tcp.keep_alive 不能小于 10 秒"))
	}
	if c.Server.TCP.ReadTimeout < 5 {
		errs = append(errs, fmt.Errorf("server.tcp.read_timeout 不能小于 5 秒"))
	}

	// V2Ray 配置
	if c.Server.V2ray.Port <= 0 || c.Server.V2ray.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.v2ray.port 必须在 1-65535 之间"))
	}

	// TLS 配置
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			errs = append(errs, fmt.Errorf("server.tls.cert_file 不能为空"))
		}
		if c.Server.TLS.KeyFile == "" {
			errs = append(errs, fmt.Errorf("server.tls.key_file 不能为空"))
		}
	}

	// 数据库配置
	if c.DB.MaxIdleConns < 1 {
		errs = append(errs, fmt.Errorf("database.max_idle_conns 不能小于 1"))
	}
	if c.DB.MaxOpenConns < 1 {
		errs = append(errs, fmt.Errorf("database.max_open_conns 不能小于 1"))
	}
	if c.DB.ConnMaxLifetime < 60 {
		errs = append(errs, fmt.Errorf("database.conn_max_lifetime 不能小于 60 秒"))
	}

	return errors.Join(errs...)
}

// Get 获取配置
func Get() *Config {
	if cfg == nil {
		panic("配置未初始化")
	}
	return cfg
}
