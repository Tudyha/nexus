package response

type DashboardResponse struct {
	SysInfo     SysInfo     `json:"sys_info"`
	ClientStats ClientStats `json:"client_stats"`
	SysStats    SysStats    `json:"sys_stats"`
}

type SysInfo struct {
	Hostname        string `json:"hostname"`
	Username        string `json:"username"`
	Uid             string `json:"uid"`
	Gid             string `json:"gid"`
	Os              string `json:"os"`
	Arch            string `json:"arch"`
	KernelArch      string `json:"kernel_arch"`
	OsName          string `json:"os_name"`
	Uptime          uint64 `json:"uptime"`
	BootTime        uint64 `json:"boot_time"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	HostId          string `json:"host_id"`
}

type ClientStats struct {
	Online  int64 `json:"online"`
	Offline int64 `json:"offline"`
}

type SysStats struct {
	CpuUsage float64 `json:"cpu_usage"`
	MemUsage float64 `json:"mem_usage"`
}
