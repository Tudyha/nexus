package sys

import (
	"github.com/jinzhu/copier"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CpuInfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
	Flags      []string `json:"flags"`
	Microcode  string   `json:"microcode"`
}

// CpuCount 获取CPU核数
func CpuCount() int {
	cpuNum, err := cpu.Counts(false)
	if err != nil {
		return 0
	}
	return cpuNum
}

// CpuInfo 获取CPU信息
func CpuInfo() ([]CpuInfoStat, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	var cpuInfoStat []CpuInfoStat
	copier.Copy(&cpuInfoStat, cpuInfo)
	return cpuInfoStat, nil
}

// CpuPercent 获取CPU使用率
func CpuPercent() (float64, error) {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}
	return cpuPercent[0], nil
}
