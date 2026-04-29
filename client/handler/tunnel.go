package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/creack/pty"
)

type TunnelHandler struct {
}

func NewTunnelHandler() conn.MessageHandler {
	return &TunnelHandler{}
}

func (t *TunnelHandler) Handle(ctx conn.Context) error {
	var err error
	var target io.ReadWriteCloser
	var req proto.TunnelOpenReq

	// 无论是否成功，都需要发送响应结果
	defer func() {
		code := proto.ErrorCode_SUCCESS
		msg := ""
		if err != nil {
			code = proto.ErrorCode_FAILED
			msg = err.Error()
		}
		ctx.GetConn().WriteMessage(proto.MessageType_TUUNEL_OPEN_ACK, &proto.Response{
			Code: code,
			Msg:  msg,
		})
	}()

	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}

	switch req.GetType() {
	case proto.TunnelType_TCP:
		target, err = t.openTcp(req.RemoteAddr)
		if err != nil {
			return err
		}
	case proto.TunnelType_UDP:
		target, err = t.openUdp(req.RemoteAddr)
		if err != nil {
			return err
		}
	case proto.TunnelType_PTY:
		target, err = t.openPty()
		if err != nil {
			return err
		}
	default:
	}

	// 劫持底层连接
	src, err := ctx.Hijack()
	if err != nil {
		return err
	}

	//TODO: 待优化
	go func() {
		defer func() {
			target.Close()
			src.Close()
		}()
		go io.Copy(target, src)
		io.Copy(src, target)
	}()
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
	return newPTYConn(ptmx, cmd), nil
}

func defaultShell() string {
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

type ptyMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

type ptyConn struct {
	ptmx *os.File
	cmd  *exec.Cmd

	mu sync.Mutex
}

func newPTYConn(ptmx *os.File, cmd *exec.Cmd) *ptyConn {
	return &ptyConn{
		ptmx: ptmx,
		cmd:  cmd,
	}
}

func (c *ptyConn) Read(p []byte) (int, error) {
	return c.ptmx.Read(p)
}

func (c *ptyConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var msg ptyMsg
	if err := json.Unmarshal(p, &msg); err != nil {
		return 0, fmt.Errorf("decode pty msg: %w", err)
	}

	switch msg.Type {
	case "resize":
		if msg.Rows == 0 || msg.Cols == 0 {
			return 0, fmt.Errorf("invalid resize: rows=%d cols=%d", msg.Rows, msg.Cols)
		}
		if err := pty.Setsize(c.ptmx, &pty.Winsize{
			Rows: msg.Rows,
			Cols: msg.Cols,
		}); err != nil {
			return 0, fmt.Errorf("setsize: %w", err)
		}

	case "data":
		if msg.Data == "" {
			return len(p), nil // 空输入正常忽略
		}
		if _, err := io.WriteString(c.ptmx, msg.Data); err != nil {
			return 0, fmt.Errorf("write stdin: %w", err)
		}

	default:
		return 0, fmt.Errorf("unknown pty msg type: %q", msg.Type)
	}
	return len(p), nil
}
func (c *ptyConn) Close() error {
	if c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
		c.cmd.Wait() // 回收进程，防止僵尸
	}
	return c.ptmx.Close()
}
