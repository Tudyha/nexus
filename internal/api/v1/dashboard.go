package v1

import (
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/sys"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

type DashboardController struct {
	clientService service.ClientService
}

func newDashboardController() *DashboardController {
	return &DashboardController{
		clientService: service.GetClientService(),
	}
}

func (c *DashboardController) GetDashboard(ctx *gin.Context) {
	var res response.DashboardResponse
	basicInfo, _ := sys.GetBasicInfo()
	if basicInfo != nil {
		copier.Copy(&res.SysInfo, basicInfo)
	}
	hostInfo, _ := sys.GetHostInfo()
	if hostInfo != nil {
		copier.Copy(&res.SysInfo, hostInfo)
	}

	online, offline, _ := c.clientService.CountByAppID(ctx.Request.Context(), getAppID(ctx))
	res.ClientStats.Online = online
	res.ClientStats.Offline = offline

	cpuPercent, _ := sys.CpuPercent()
	res.SysStats.CpuUsage = cpuPercent

	res.SysStats.MemUsage = sys.MemUsage()

	response.Success(ctx, res)
}
