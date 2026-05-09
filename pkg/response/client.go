package response

import (
	"time"

	"github.com/Tudyha/nexus/pkg/enum"
)

type ClientResponse struct {
	ID              uint64            `json:"id"`
	AppID           uint64            `json:"app_id"`
	Status          enum.ClientStatus `json:"status"`
	LastOnlineTime  time.Time         `json:"last_online_time"`
	CreatedAt       *time.Time        `json:"created_at"`
	UpdatedAt       *time.Time        `json:"updated_at"`
	Version         uint32            `json:"version"`
	VersionName     string            `json:"version_name"`
	Hostname        string            `json:"hostname"`
	Username        string            `json:"username"`
	Gid             string            `json:"gid"`
	RemoteIP        string            `json:"remote_ip"`
	RemoteIpCountry string            `json:"remote_ip_country"`
	LocalIP         string            `json:"local_ip"`
	Port            string            `json:"port"`
	Uptime          uint64            `json:"uptime"`
	BootTime        uint64            `json:"boot_time"`
	Platform        string            `json:"platform"`
	PlatformFamily  string            `json:"platform_family"`
	PlatformVersion string            `json:"platform_version"`
	KernelVersion   string            `json:"kernel_version"`
	KernelArch      string            `json:"kernel_arch"`
	CpuInfo         string            `json:"cpu_info"`
	MemTotal        uint64            `json:"mem_total"`
	DiskTotal       uint64            `json:"disk_total"`
	DeviceId        string            `json:"device_id"` // 设备id
	Os              string            `json:"os"`        // 操作系统
	Arch            string            `json:"arch"`      // 操作系统架构
	HostId          string            `json:"host_id"`   // 主机id
	Uid             string            `json:"uid"`       // 用户id
	CpuCount        int               `json:"cpu_count"` // cpu核数
	Task            *TaskExecutionResponse `json:"task"`
}

type ClientBindResponse struct {
	MacBind     string `json:"mac_bind"`
	WindowsBind string `json:"windows_bind"`
	LinuxBind   string `json:"linux_bind"`
}
