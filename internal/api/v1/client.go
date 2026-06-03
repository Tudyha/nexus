package v1

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Tudyha/nexus/internal/config"
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/errcode"
	nexusio "github.com/Tudyha/nexus/pkg/io"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/copier"
	pb "google.golang.org/protobuf/proto"
)

type ClientController struct {
	clientService  service.ClientService
	appService     service.AppService
	versionService service.VersionService
	sessionManager session.Manager
	taskService    service.TaskService
}

func newClientController() *ClientController {
	return &ClientController{
		clientService:  service.GetClientService(),
		appService:     service.GetAppService(),
		versionService: service.GetVersionService(),
		sessionManager: session.GetManager(),
		taskService:    service.GetTaskService(),
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

	type clientConfig struct {
		ServerAddr string `json:"server_addr"` // 服务器地址
		AppId      int64  `json:"app_id"`      // 应用id
		AppSecret  string `json:"app_secret"`  // 应用密钥
	}

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

	// 查询各平台最新版本
	type osArch struct{ os, arch string }
	targets := map[string]osArch{
		"mac":   {"darwin", "amd64"},
		"linux": {"linux", "amd64"},
		"win":   {"windows", "amd64"},
	}

	latestMap := make(map[string]string)
	for key, t := range targets {
		v, err := h.versionService.GetLatestByOS(ctx, t.os, t.arch)
		if err == nil && v != nil {
			latestMap[key] = strings.TrimPrefix(v.BinaryPath, "tmp/")
		}
	}

	sh := "curl -s http://%s:%d/%s -o /tmp/nexus-cli && chmod +x /tmp/nexus-cli && /tmp/nexus-cli run -d -c %s"
	res := response.ClientBindResponse{
		MacBind:     fmt.Sprintf(sh, cfg.Server.Host, cfg.Server.HTTP.Port, latestMap["mac"], c),
		LinuxBind:   fmt.Sprintf(sh, cfg.Server.Host, cfg.Server.HTTP.Port, latestMap["linux"], c),
		WindowsBind: fmt.Sprintf(sh, cfg.Server.Host, cfg.Server.HTTP.Port, latestMap["win"], c),
	}
	response.Success(ctx, res)
}

// 更新客户端配置
func (h *ClientController) UpdateConfig(ctx *gin.Context) {
	appID := getAppID(ctx)
	if appID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	var req struct {
		Config string `json:"config"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}
	if err := h.appService.UpdateConfig(ctx, appID, req.Config); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// 获取应用配置
func (h *ClientController) GetConfig(ctx *gin.Context) {
	appID := getAppID(ctx)
	if appID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	app, err := h.appService.GetApp(ctx, appID)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, gin.H{"config": app.Config})
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
	response.Success(ctx, res)
}

// 获取在线客户端列表
func (h *ClientController) ListOnline(ctx *gin.Context) {
	appID := getAppID(ctx)
	if appID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	clients, err := h.clientService.ListOnline(ctx, appID)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	type simpleClient struct {
		ID       uint64 `json:"id"`
		Hostname string `json:"hostname"`
	}
	var list []simpleClient
	for _, c := range clients {
		list = append(list, simpleClient{ID: c.ID, Hostname: c.Hostname})
	}
	response.Success(ctx, list)
}

// 删除客户端
func (h *ClientController) Delete(ctx *gin.Context) {
	if s, err := h.getSession(ctx); s != nil && err == nil {
		s.Exit()
	}

	response.Success(ctx, nil)
}

func (h *ClientController) getSession(ctx *gin.Context) (*session.Session, error) {
	client, err := h.clientService.GetByID(ctx, getClientID(ctx))
	if err != nil {
		return nil, err
	}
	return h.sessionManager.GetSession(client.SessionID)
}

// 在线终端
func (h *ClientController) Terminal(ctx *gin.Context) {
	s, err := h.getSession(ctx)
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

// 客户端升级
func (h *ClientController) Upgrade(ctx *gin.Context) {
	clientID := getClientID(ctx)
	if clientID == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	force := ctx.DefaultQuery("force", "false") == "true"

	latest, err := h.versionService.GetLatestForClient(ctx, clientID)
	if err != nil {
		response.Fail(ctx, err)
		return
	}

	sess, err := h.getSession(ctx)
	if err != nil {
		response.Fail(ctx, errcode.ErrVersionNotReady)
		return
	}

	// 创建任务
	execs, err := h.taskService.CreateTask(ctx, 1, []uint64{clientID}) // 1: UPGRADE
	if err != nil {
		response.Fail(ctx, errcode.ErrVersionUpgrade)
		return
	}
	exec := execs[0]

	// 构造下载地址（拉取升级）
	cfg := config.Get()
	downloadURL := fmt.Sprintf("http://%s:%d/%s",
		cfg.Server.Host, cfg.Server.HTTP.Port, strings.Replace(latest.BinaryPath, "tmp/", "", -1))

	// 返回 task_id 给前端
	response.Success(ctx, gin.H{"task_id": exec.TaskID})

	// goroutine 异步执行
	go func() {
		payload := &proto.UpgradePayload{
			Version:     latest.Version,
			VersionName: latest.VersionName,
			Checksum:    latest.Checksum,
			BinarySize:  latest.BinarySize,
			DownloadUrl: downloadURL,
			Force:       force,
		}
		b, _ := pb.Marshal(payload)

		done := false

		if err := sess.SendTask(exec.ID, proto.TaskType_UPGRADE, b, true, func(p *proto.TaskProgress) {
			if p.Done {
				done = true
			}
			h.taskService.UpdateProgress(ctx, exec.ID, p)
		}); err != nil {
			// 任务执行失败
			h.taskService.UpdateProgress(ctx, exec.ID, &proto.TaskProgress{
				Message: err.Error(),
				Done:    true,
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		if !done {
			// 如果任务未完成就退出了，标记为中断
			h.taskService.UpdateExecution(ctx, &model.TaskExecution{
				BaseModel: model.BaseModel{ID: exec.ID},
				Status:    enum.TaskStatusInterrupt,
				Error:     "task interrupted",
			})
		}
	}()
}

// sendAndRecv 通过 smux 发送消息并读取响应
func (h *ClientController) sendAndRecv(ctx *gin.Context, msgType proto.MessageType, req, resp any) error {
	s, err := h.getSession(ctx)
	if err != nil {
		return err
	}
	return s.Command(msgType, req, resp)
}

// ---- 进程管理 ----

func (h *ClientController) ProcessList(ctx *gin.Context) {
	var resp proto.ProcessListResp
	if err := h.sendAndRecv(ctx, proto.MessageType_PROCESS_LIST, &proto.ProcessListReq{}, &resp); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, resp.List)
}

func (h *ClientController) ProcessKill(ctx *gin.Context) {
	pid := utils.StringToUint64(ctx.Param("pid"))
	if pid == 0 {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, "invalid pid")
		return
	}

	if err := h.sendAndRecv(ctx, proto.MessageType_PROCESS_KILL, &proto.ProcessKillReq{Pid: int32(pid)}, nil); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// ---- 网络管理 ----

func (h *ClientController) NetworkList(ctx *gin.Context) {
	var resp proto.NetworkListResp
	if err := h.sendAndRecv(ctx, proto.MessageType_NETWORK_LIST, &proto.NetworkListReq{}, &resp); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, resp.List)
}

// ---- 文件管理 ----

func (h *ClientController) FileList(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		path = "/"
	}

	var resp proto.FileListResp
	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_LIST, &proto.FileListReq{Path: path}, &resp); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, resp.Entries)
}

func (h *ClientController) FileDownload(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		response.FailWithMsg(ctx, nil, "path required")
		return
	}

	var resp proto.FileDownloadResp
	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_DOWNLOAD, &proto.FileDownloadReq{Path: path}, &resp); err != nil {
		response.Fail(ctx, err)
		return
	}
	ctx.Data(http.StatusOK, "application/octet-stream", resp.Data)
}

func (h *ClientController) FileUpload(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		response.FailWithMsg(ctx, nil, "path required")
		return
	}
	data, err := ctx.GetRawData()
	if err != nil {
		response.FailWithMsg(ctx, nil, "read body failed")
		return
	}

	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_UPLOAD, &proto.FileUploadReq{
		Path:   path,
		Data:   data,
		Append: ctx.Query("append") == "true",
	}, nil); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

func (h *ClientController) FileDelete(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		response.FailWithMsg(ctx, nil, "path required")
		return
	}

	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_DELETE, &proto.FileDeleteReq{Path: path}, nil); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

func (h *ClientController) FileMkdir(ctx *gin.Context) {
	path := ctx.Query("path")
	if path == "" {
		response.FailWithMsg(ctx, nil, "path required")
		return
	}

	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_MKDIR, &proto.FileMkdirReq{Path: path}, nil); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

func (h *ClientController) FileRename(ctx *gin.Context) {
	var req struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.OldPath == "" || req.NewPath == "" {
		response.FailWithMsg(ctx, nil, "old_path and new_path required")
		return
	}

	if err := h.sendAndRecv(ctx, proto.MessageType_FILE_RENAME, &proto.FileRenameReq{OldPath: req.OldPath, NewPath: req.NewPath}, nil); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}
