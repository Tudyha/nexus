package model

import (
	"time"

	"github.com/Tudyha/nexus/pkg/enum"
)

type Client struct {
	BaseModel

	// 应用信息
	AppID          uint64            `gorm:"column:app_id;not null"`     // 应用id
	Version        uint32            `gorm:"column:version"`             // client版本号
		VersionName    string            `gorm:"column:version_name"`        // client版本名称
	SessionID      string            `gorm:"column:session_id;not null"` // 连接id
	Status         enum.ClientStatus `gorm:"column:status"`              // 连接状态
	LastOnlineTime time.Time         `gorm:"column:last_online_time"`    // 最后一次连接时间

	// 设备信息
	DeviceId        string `gorm:"column:device_id;uniqueIndex"` // 设备id
	Os              string `gorm:"column:os"`                    // 操作系统
	Arch            string `gorm:"column:arch"`                  // 操作系统架构
	Hostname        string `gorm:"column:hostname"`              // 主机名
	HostId          string `gorm:"column:host_id"`               // 主机id
	Username        string `gorm:"column:username"`              // 用户名
	Gid             string `gorm:"column:gid"`                   // 组id
	Uid             string `gorm:"column:uid"`                   // 用户id
	RemoteIP        string `gorm:"column:remote_ip;not null"`    // 远程IP
	RemoteIpCountry string `gorm:"column:remote_ip_country"`     // 远程IP所属国家
	LocalIP         string `gorm:"column:local_ip"`              // 本地IP
	Port            string `gorm:"column:port"`                  // 端口
	Uptime          uint64 `gorm:"column:uptime"`                // 运行时间
	BootTime        uint64 `gorm:"column:boot_time"`             // 启动时间
	Platform        string `gorm:"column:platform"`              // 平台名称
	PlatformFamily  string `gorm:"column:platform_family"`       // 平台
	PlatformVersion string `gorm:"column:platform_version"`      // 平台版本
	KernelVersion   string `gorm:"column:kernel_version"`        // 内核版本
	KernelArch      string `gorm:"column:kernel_arch"`           // 内核架构
	CpuCount        int    `gorm:"column:cpu_count"`             // cpu核数
	CpuInfo         string `gorm:"column:cpu_info;type:json"`    // cpu信息
	MemTotal        uint64 `gorm:"column:mem_total"`             // 内存总大小
	DiskTotal       uint64 `gorm:"column:disk_total"`            // 磁盘总大小
}

func (Client) TableName() string {
	return "t_client"
}

type ClientStat struct {
	BaseModel
	ClientID uint64 `gorm:"column:client_id;not null;index"` // 客户端id

	CpuPercent float64 // cpu使用率

	MemTotal       uint64  // 内存总大小
	MemUsed        uint64  // 内存使用大小
	MemFree        uint64  // 内存剩余大小
	MemCached      uint64  // 内存缓存
	MemBuffers     uint64  // 内存缓冲
	MemAvailable   uint64  // 可用内存
	MemUsedPercent float64 // 内存使用率
	MemSwapTotal   uint64  // 内存交换分区总大小
	MemSwapUsed    uint64  // 内存交换分区使用大小
	MemSwapFree    uint64  // 内存交换分区剩余大小

	NetBytesSent float64 // 网卡发送字节数
	NetBytesRecv float64 // 网卡接收字节数
}

func (ClientStat) TableName() string {
	return "t_client_stat"
}
