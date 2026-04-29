package sys

import (
	"net"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/host"
)

type BasicInfo struct {
	Hostname string
	Username string
	Uid      string
	Gid      string
	Os       string
	Arch     string
}

type HostInfo struct {
	Uptime          uint64
	BootTime        uint64
	Platform        string
	PlatformFamily  string
	PlatformVersion string
	KernelVersion   string
	HostId          string
	KernelArch      string
}

// 获取基本信息，不使用gopsutil, 不会有兼容性问题
func GetBasicInfo() (*BasicInfo, error) {
	info := &BasicInfo{}
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		return info, err
	}
	info.Hostname = hostname

	// 获取当前用户
	currentUser, err := user.Current()
	if err != nil {
		return info, err
	}
	info.Username = currentUser.Username
	info.Uid = currentUser.Uid
	info.Gid = currentUser.Gid

	// 操作系统和架构
	info.Os = runtime.GOOS
	info.Arch = runtime.GOARCH
	return info, nil
}

// 获取主机信息, 使用gopsutil
func GetHostInfo() (*HostInfo, error) {
	info := &HostInfo{}

	hostInfo, err := host.Info()
	if err != nil {
		return info, err
	}
	// 填充主机信息
	info.Platform = hostInfo.Platform
	info.PlatformFamily = hostInfo.PlatformFamily
	info.PlatformVersion = hostInfo.PlatformVersion
	info.KernelVersion = hostInfo.KernelVersion
	info.HostId = hostInfo.HostID
	info.KernelArch = hostInfo.KernelArch

	// 启动时间相关
	info.BootTime = uint64(hostInfo.BootTime)
	if hostInfo.Uptime > 0 {
		info.Uptime = uint64(hostInfo.Uptime)
	} else {
		info.Uptime = uint64(time.Now().Unix()) - hostInfo.BootTime
	}
	return info, nil
}

func GetDeviceId() string {
	hostInfo, err := host.Info()
	if err != nil {
		return GetMacAddress()
	}
	return hostInfo.HostID
}

func GetMacAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var address []string
	for _, i := range interfaces {
		a := i.HardwareAddr.String()
		if a != "" {
			address = append(address, a)
		}
	}
	if len(address) == 0 {
		return ""
	}
	return address[0]
}
