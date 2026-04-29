package config

import (
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
	Env   string      `mapstructure:"env"`   // 环境: dev, prod
	Host  string      `mapstructure:"host"`  // 服务监听地址
	HTTP  HttpConfig  `mapstructure:"http"`  // http服务配置
	TCP   TCPConfig   `mapstructure:"tcp"`   // tcp服务配置
	V2ray V2rayConfig `mapstructure:"v2ray"` // v2ray服务配置
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

	return nil
}

// Get 获取配置
func Get() *Config {
	if cfg == nil {
		panic("配置未初始化")
	}
	return cfg
}
