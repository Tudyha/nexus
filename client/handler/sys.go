package handler

import (
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/sys"
	"github.com/jinzhu/copier"
)

type SysHandler struct {
}

func NewSysHandler() *SysHandler {
	return &SysHandler{}
}

func (h *SysHandler) Info() (*proto.ClientInfo, error) {
	var info proto.ClientInfo
	info.DeviceId = sys.GetDeviceId()
	baseInfo, err := sys.GetBasicInfo()
	if err != nil {
		return nil, err
	}
	copier.Copy(&info, baseInfo)

	hostInfo, err := sys.GetHostInfo()
	if err != nil {
		return nil, err
	}
	copier.Copy(&info, hostInfo)

	info.CpuCount = uint32(sys.CpuCount())
	info.MemTotal = sys.MemoryTotal()
	info.DiskTotal = sys.DiskTotal()

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
	copier.Copy(&stats.MemStat, memStat)

	bytesSent, bytesRecv := sys.NetSpeed()
	stats.NetStat.BytesSent = bytesSent
	stats.NetStat.BytesRecv = bytesRecv

	return stats, nil
}
