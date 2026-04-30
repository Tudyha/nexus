package io

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

type PtyReadWriteCloser struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

func NewPtyReadWriteCloser(ptmx *os.File, cmd *exec.Cmd) *PtyReadWriteCloser {
	return &PtyReadWriteCloser{
		ptmx: ptmx,
		cmd:  cmd,
	}
}

func (c *PtyReadWriteCloser) Read(p []byte) (int, error) {
	return c.ptmx.Read(p)
}

func (c *PtyReadWriteCloser) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	switch p[0] {
	case 0x00: // 数据帧
		if len(p) == 1 {
			return 1, nil
		}
		if _, err := c.ptmx.Write(p[1:]); err != nil {
			return 0, err
		}
	case 0x01: // resize帧
		if len(p) != 5 {
			return 0, fmt.Errorf("invalid resize frame length: %d", len(p))
		}
		rows := binary.BigEndian.Uint16(p[1:3])
		cols := binary.BigEndian.Uint16(p[3:5])
		if rows == 0 || cols == 0 {
			return 0, fmt.Errorf("invalid resize: rows=%d cols=%d", rows, cols)
		}
		if err := pty.Setsize(c.ptmx, &pty.Winsize{
			Rows: rows,
			Cols: cols,
		}); err != nil {
			return 0, err
		}
	default:
		return 0, fmt.Errorf("unknown frame type: 0x%02x", p[0])
	}

	return len(p), nil
}

func (c *PtyReadWriteCloser) Close() error {
	if c.cmd.Process != nil {
		c.cmd.Process.Signal(syscall.SIGTERM)
		c.cmd.Wait() // 回收进程，防止僵尸
	}
	return c.ptmx.Close()
}
