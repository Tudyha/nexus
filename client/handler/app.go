package handler

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

// appTemplate 应用模板定义
type appTemplate struct {
	ID        string   // 唯一标识
	Name      string   // 显示名称
	Category  string   // 分类
	Icon      string   // 图标名
	CheckCmd  string   // 检测命令
	CheckArgs []string // 检测命令参数
	InstallCmd string  // 安装命令（通过 bash -c 执行）
}

// appTemplates 预置的应用商店列表
var appTemplates = []appTemplate{
	{ID: "docker", Name: "Docker", Category: "runtime", Icon: "mdi:docker",
		CheckCmd: "docker", CheckArgs: []string{"--version"},
		InstallCmd: "curl -fsSL https://get.docker.com | sh"},
	{ID: "mysql", Name: "MySQL", Category: "database", Icon: "mdi:database",
		CheckCmd: "mysql", CheckArgs: []string{"--version"},
		InstallCmd: "apt-get install -y mysql-server 2>/dev/null || yum install -y mysql-server 2>/dev/null || apk add mysql"},
	{ID: "nginx", Name: "Nginx", Category: "web", Icon: "mdi:nginx",
		CheckCmd: "nginx", CheckArgs: []string{"-v"},
		InstallCmd: "apt-get install -y nginx 2>/dev/null || yum install -y nginx 2>/dev/null || apk add nginx"},
	{ID: "redis", Name: "Redis", Category: "database", Icon: "mdi:redis",
		CheckCmd: "redis-cli", CheckArgs: []string{"--version"},
		InstallCmd: "apt-get install -y redis-server 2>/dev/null || yum install -y redis 2>/dev/null || apk add redis"},
	{ID: "node", Name: "Node.js", Category: "runtime", Icon: "mdi:nodejs",
		CheckCmd: "node", CheckArgs: []string{"--version"},
		InstallCmd: "curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs"},
	{ID: "python", Name: "Python", Category: "runtime", Icon: "mdi:language-python",
		CheckCmd: "python3", CheckArgs: []string{"--version"},
		InstallCmd: "apt-get install -y python3 2>/dev/null || yum install -y python3 2>/dev/null || apk add python3"},
}

// getTemplate 根据 ID 查找应用模板
func getTemplate(id string) *appTemplate {
	for _, t := range appTemplates {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

// ---- APP_CHECK ----

type AppCheckHandler struct{}

func NewAppCheckHandler() conn.MessageHandler { return &AppCheckHandler{} }
func (h *AppCheckHandler) Type() proto.MessageType { return proto.MessageType_APP_CHECK }
func (h *AppCheckHandler) Handle(ctx conn.Context) error {
	var req proto.AppCheckReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	log.Info().Int("app_count", len(req.AppIds)).Msg("app check")

	want := make(map[string]bool)
	for _, id := range req.AppIds {
		want[id] = true
	}

	var resp proto.AppCheckResp
	for _, tmpl := range appTemplates {
		if len(want) > 0 && !want[tmpl.ID] {
			continue
		}
		entry := &proto.AppEntry{
			AppId:    tmpl.ID,
			Name:     tmpl.Name,
			Category: tmpl.Category,
			Icon:     tmpl.Icon,
		}
		if ver := detectVersion(tmpl.CheckCmd, tmpl.CheckArgs); ver != "" {
			entry.Version = ver
		}
		resp.Entries = append(resp.Entries, entry)
	}

	return ctx.GetConn().WriteMessage(proto.MessageType_APP_CHECK, &resp)
}

// ---- APP_INSTALL ----

type AppInstallHandler struct{}

func NewAppInstallHandler() conn.MessageHandler { return &AppInstallHandler{} }
func (h *AppInstallHandler) Type() proto.MessageType { return proto.MessageType_APP_INSTALL }
func (h *AppInstallHandler) Handle(ctx conn.Context) error {
	var req proto.AppInstallReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	tmpl := getTemplate(req.AppId)
	if tmpl == nil {
		return ctx.GetConn().WriteMessage(proto.MessageType_APP_INSTALL, &proto.Response{Code: 2})
	}

	log.Info().Str("app", tmpl.ID).Msg("installing app")
	cmd := exec.Command("/bin/sh", "-c", tmpl.InstallCmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Error().Err(err).Str("output", stderr.String()).Str("app", tmpl.ID).Msg("app install failed")
		return ctx.GetConn().WriteMessage(proto.MessageType_APP_INSTALL, &proto.Response{
			Code: 2,
			Msg:  err.Error() + ": " + stderr.String(),
		})
	}
	log.Info().Str("app", tmpl.ID).Msg("app installed successfully")
	return ctx.GetConn().WriteMessage(proto.MessageType_APP_INSTALL, &proto.Response{Code: 1, Msg: stdout.String()})
}

// ---- helpers ----

func detectVersion(cmd string, args []string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		log.Debug().Str("cmd", cmd).Err(err).Msg("app check failed")
		return ""
	}
	ver := strings.TrimSpace(string(out))
	if idx := strings.IndexByte(ver, '\n'); idx > 0 {
		ver = ver[:idx]
	}
	return ver
}
