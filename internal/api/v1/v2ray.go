package v1

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
)

type V2RayController struct {
	clientService service.ClientService
}

func newV2RayController() *V2RayController {
	return &V2RayController{
		clientService: service.GetClientService(),
	}
}

func (h *V2RayController) V2raySubscribe(ctx *gin.Context) {
	v := ctx.Param("appId")
	appId := utils.StringToUint64(v)
	clients, err := h.clientService.ListOnline(ctx, appId)
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	cfg := config.Get()
	links := []string{}

	host, port := cfg.Server.Host, cfg.Server.V2ray.Port

	for _, agent := range clients {
		addr := fmt.Sprintf(`{"add":"%s","id":"%s","net":"tcp","port":"%d","ps":"%s:%d","scy":"auto","type":"none","v":"2"}`,
			host, agent.SessionID, port, host, port)
		links = append(links, "vmess://"+base64.StdEncoding.EncodeToString([]byte(addr)))
	}

	content := strings.Join(links, "\n")
	ctx.String(http.StatusOK, content)
}
