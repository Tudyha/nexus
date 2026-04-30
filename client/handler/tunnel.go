package handler

import (
	"bufio"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"

	"github.com/Tudyha/nexus/pkg/conn"
	nexusio "github.com/Tudyha/nexus/pkg/io"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/creack/pty"
)

var (
	cachedShell string
	shellOnce   sync.Once
)

type TunnelHandler struct {
}

func NewTunnelHandler() conn.MessageHandler {
	return &TunnelHandler{}
}

func (t *TunnelHandler) Handle(ctx conn.Context) error {
	var err error
	var dst io.ReadWriteCloser
	var req proto.TunnelOpenReq

	// 无论是否成功，都需要发送响应结果
	defer func() {
		code := proto.ErrorCode_SUCCESS
		msg := ""
		if err != nil {
			code = proto.ErrorCode_FAILED
			msg = err.Error()
		}
		ctx.GetConn().WriteMessage(proto.MessageType_TUNNEL_OPEN_ACK, &proto.Response{
			Code: code,
			Msg:  msg,
		})
	}()

	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	switch req.GetType() {
	case proto.TunnelType_TCP:
		dst, err = t.openTcp(req.RemoteAddr)
		if err != nil {
			return err
		}
	case proto.TunnelType_UDP:
		dst, err = t.openUdp(req.RemoteAddr)
		if err != nil {
			return err
		}
	case proto.TunnelType_PTY:
		dst, err = t.openPty()
		if err != nil {
			return err
		}
	default:
	}

	// 劫持底层连接，上层不会再进行数据读写
	src, err := ctx.Hijack()
	if err != nil {
		return err
	}

	// 数据转发
	go nexusio.Copy(src, dst)
	return nil
}

func (t *TunnelHandler) Type() proto.MessageType {
	return proto.MessageType_TUNNEL_OPEN
}

func (p *TunnelHandler) openTcp(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", addr)
}

func (p *TunnelHandler) openUdp(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

func (p *TunnelHandler) openPty() (io.ReadWriteCloser, error) {
	shell := defaultShell()
	cmd := exec.Command(shell)
	// 继承环境，让 shell 正常工作
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return nexusio.NewPtyReadWriteCloser(ptmx, cmd), nil
}

func defaultShell() string {
	shellOnce.Do(func() {
		cachedShell = resolveShell()
	})
	return cachedShell
}

func resolveShell() string {
	// 优先读环境变量（尊重用户配置）
	if s := os.Getenv("SHELL"); s != "" {
		return s
	}
	// Windows
	if runtime.GOOS == "windows" {
		if s := os.Getenv("COMSPEC"); s != "" {
			return s
		}
		return "cmd.exe"
	}

	if runtime.GOOS == "darwin" {
		if shell := shellFromDscl(); shell != "" {
			return shell
		}
	}

	if shell := shellFromPasswd(); shell != "" {
		return shell
	}
	return "/bin/sh" // 最终兜底，所有 POSIX 系统都有
}

func shellFromPasswd() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	prefix := u.Username + ":"
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		// root:x:0:0:root:/root:/bin/bash
		//  0   1 2 3  4     5      6
		parts := strings.SplitN(line, ":", 7)
		if len(parts) == 7 {
			return parts[6]
		}
	}
	return ""
}

func shellFromDscl() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	out, err := exec.Command("dscl", ".", "-read",
		"/Users/"+u.Username, "UserShell").Output()
	if err != nil {
		return ""
	}
	// 输出格式: "UserShell: /bin/zsh"
	parts := strings.SplitN(strings.TrimSpace(string(out)), " ", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
