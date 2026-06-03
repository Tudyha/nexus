package handler

import (
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/sys"
)

// SysHandler 收集系统信息用于握手和心跳。
// 设备 ID、CPU 核数、内存总量、磁盘总量在进程生命周期内不变，启动时缓存一次。
type SysHandler struct {
	deviceId  string
	cpuCount  uint32
	memTotal  uint64
	diskTotal uint64
}

func NewSysHandler() *SysHandler {
	return &SysHandler{
		deviceId:  sys.GetDeviceId(),
		cpuCount:  uint32(sys.CpuCount()),
		memTotal:  sys.MemoryTotal(),
		diskTotal: sys.DiskTotal(),
	}
}

func (h *SysHandler) Info() (*proto.ClientInfo, error) {
	var info proto.ClientInfo
	info.DeviceId = h.deviceId
	info.CpuCount = h.cpuCount
	info.MemTotal = h.memTotal
	info.DiskTotal = h.diskTotal

	baseInfo, err := sys.GetBasicInfo()
	if err != nil {
		return nil, err
	}
	info.Hostname = baseInfo.Hostname
	info.Username = baseInfo.Username
	info.Uid = baseInfo.Uid
	info.Gid = baseInfo.Gid
	info.Os = baseInfo.Os
	info.Arch = baseInfo.Arch

	hostInfo, err := sys.GetHostInfo()
	if err != nil {
		return nil, err
	}
	info.Uptime = hostInfo.Uptime
	info.BootTime = hostInfo.BootTime
	info.Platform = hostInfo.Platform
	info.PlatformFamily = hostInfo.PlatformFamily
	info.PlatformVersion = hostInfo.PlatformVersion
	info.KernelVersion = hostInfo.KernelVersion
	info.HostId = hostInfo.HostId
	info.KernelArch = hostInfo.KernelArch

	return &info, nil
}

func (h *SysHandler) SystemStats() (*proto.HeartbeatReq, error) {
	stats := &proto.HeartbeatReq{
		CpuStat:  &proto.CpuStat{},
		MemStat:  &proto.MemStat{},
		NetStat:  &proto.NetStat{},
		DiskStat: &proto.DiskStat{},
	}

	cpuPercent, err := sys.CpuPercent()
	if err != nil {
		return nil, err
	}
	stats.CpuStat.Percent = cpuPercent

	memStat, err := sys.MemoryStat()
	if err != nil {
		return nil, err
	}
	stats.MemStat.Total = memStat.Total
	stats.MemStat.Used = memStat.Used
	stats.MemStat.Free = memStat.Free
	stats.MemStat.Available = memStat.Available
	stats.MemStat.UsedPercent = memStat.UsedPercent

	bytesSent, bytesRecv := sys.NetSpeed()
	stats.NetStat.BytesSent = bytesSent
	stats.NetStat.BytesRecv = bytesRecv

	return stats, nil
}
