package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	"github.com/Tudyha/nexus/pkg/errcode"
	nexusio "github.com/Tudyha/nexus/pkg/io"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
)

type ClientController struct {
	clientService  service.ClientService
	appService     service.AppService
	sessionManager session.Manager
}

func newClientController() *ClientController {
	return &ClientController{
		clientService:  service.GetClientService(),
		appService:     service.GetAppService(),
		sessionManager: session.GetManager(),
	}
}

// 获取客户端列表
func (h *ClientController) GetPage(ctx *gin.Context) {
	appID := getAppID(ctx)
	if appID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	var req request.ClientQueryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	res, err := h.clientService.GetPage(ctx, appID, &req)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, res)
}

type clientConfig struct {
	ServerAddr string `json:"server_addr"` // 服务器地址
	AppId      int64  `json:"app_id"`      // 应用id
	AppSecret  string `json:"app_secret"`  // 应用密钥
}

// 获取客户端绑定命令
func (h *ClientController) GetBind(ctx *gin.Context) {
	appID := getAppID(ctx)
	if appID == 0 {
		response.Fail(ctx, errcode.ErrAppNotFound)
		return
	}

	app, err := h.appService.GetApp(ctx, appID)
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	cfg := config.Get()

	addr := net.JoinHostPort(cfg.Server.Host, fmt.Sprintf("%d", cfg.Server.TCP.Port))
	agentConfig := clientConfig{
		ServerAddr: addr,
		AppId:      int64(app.ID),
		AppSecret:  app.AppSecret,
	}

	data, err := json.Marshal(agentConfig)
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	c := utils.Base64Encode(data)

	sh := "curl -s http://%s:%d/%s -o /tmp/nexus-cli && chmod +x /tmp/nexus-cli && /tmp/nexus-cli run -d -c %s"
	response.Success(ctx, response.ClientBindResponse{
		MacBind:   fmt.Sprintf(sh, cfg.Server.Host, cfg.Server.HTTP.Port, "nexus-cli-darwin-amd64", c),
		LinuxBind: fmt.Sprintf(sh, cfg.Server.Host, cfg.Server.HTTP.Port, "nexus-cli-linux-amd64", c),
	})
}

// 获取客户端详情
func (h *ClientController) GetByID(ctx *gin.Context) {
	client, err := h.clientService.GetByID(ctx, getClientID(ctx))
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	var res response.ClientResponse
	copier.Copy(&res, client)
	res.VersionName = "v1.0.0"
	response.Success(ctx, res)
}

// 删除客户端
func (h *ClientController) Delete(ctx *gin.Context) {
	_, err := h.clientService.DeleteByID(ctx, getClientID(ctx))
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// pty伪终端
func (h *ClientController) OpenPty(ctx *gin.Context) {
	client, err := h.clientService.GetByID(ctx, getClientID(ctx))
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	s, err := h.sessionManager.GetSession(client.SessionID)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	src, err := s.OpenTunnel(proto.TunnelType_PTY, "")
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	maxMessageSize := 32 * 1024
	upgrader := websocket.Upgrader{
		ReadBufferSize:  maxMessageSize,
		WriteBufferSize: maxMessageSize,
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	dst, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		response.Fail(ctx, err)
		src.Close()
		return
	}
	go nexusio.Copy(src, &nexusio.WebSocketReadWriteCloser{Conn: dst})
}

func (h *ClientController) GenerateV2raySubscribeLink(ctx *gin.Context) {
	var req request.GenerateV2raySubscribeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	clients, err := h.clientService.GetByIDs(ctx, req.Ids)
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
	filename := fmt.Sprintf("v2ray-sub-%s.txt", utils.MD5(content))
	res := fmt.Sprintf("http://%s:%d/%s", host, cfg.Server.HTTP.Port, filename)
	if utils.FileExists("./tmp/" + filename) {
		response.Success(ctx, res)
		return
	}

	file, err := os.Create("./tmp/" + filename)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	defer file.Close()
	_, err = file.WriteString(content)
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	response.Success(ctx, res)
}
