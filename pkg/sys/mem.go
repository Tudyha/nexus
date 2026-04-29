package sys

import "github.com/shirou/gopsutil/v4/mem"

// 总内存
func MemoryTotal() uint64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return memInfo.Total
}

func MemoryStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

// 内存使用率
func MemUsage() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return memInfo.UsedPercent
}
