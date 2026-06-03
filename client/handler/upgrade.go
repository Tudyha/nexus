package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/Tudyha/nexus/pkg/utils"
)

func executeUpgrade(c *conn.Conn, task *proto.Task, payload *proto.UpgradePayload) error {
	reportProgress(c, task.TaskId, 0, "开始升级")

	exe, err := os.Executable()
	if err != nil {
		reportProgressDone(c, task.TaskId, false, fmt.Sprintf("get executable: %v", err))
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		exe, _ = os.Executable()
	}

	tmpFile, err := os.CreateTemp("", "nexus-upgrade-*")
	if err != nil {
		reportProgressDone(c, task.TaskId, false, fmt.Sprintf("create temp file: %v", err))
		return err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if payload.DownloadUrl == "" {
		reportProgressDone(c, task.TaskId, false, "缺少下载地址")
		return fmt.Errorf("missing download url")
	}

	reportProgress(c, task.TaskId, 10, "开始下载")
	if err := downloadBinary(payload.DownloadUrl, tmpPath, func(done, total int64) {
		if total > 0 {
			pct := int32(10 + (done * 60 / total))
			if pct > 70 {
				pct = 70
			}
			reportProgress(c, task.TaskId, pct, fmt.Sprintf("下载中 %d%%", pct))
		}
	}); err != nil {
		reportProgressDone(c, task.TaskId, false, fmt.Sprintf("download: %v", err))
		return err
	}

	reportProgress(c, task.TaskId, 80, "校验文件中")
	if err := utils.VerifyChecksum(tmpPath, payload.Checksum); err != nil {
		reportProgressDone(c, task.TaskId, false, fmt.Sprintf("checksum: %v", err))
		return err
	}

	reportProgress(c, task.TaskId, 90, "替换二进制文件")
	if err := os.Rename(tmpPath, exe); err != nil {
		if err := copyAndReplace(tmpPath, exe); err != nil {
			reportProgressDone(c, task.TaskId, false, fmt.Sprintf("replace binary: %v", err))
			return err
		}
	}

	if err := os.Chmod(exe, 0755); err != nil {
		reportProgressDone(c, task.TaskId, false, fmt.Sprintf("chmod: %v", err))
		return err
	}

	reportProgress(c, task.TaskId, 100, "升级成功")
	reportProgressDone(c, task.TaskId, true, "升级成功")

	return nil
}

func reportProgress(c *conn.Conn, taskID uint64, progress int32, message string) {
	c.WriteMessage(proto.MessageType_TASK_PROGRESS, &proto.TaskProgress{
		TaskId:   taskID,
		Progress: progress,
		Message:  message,
	})
}

func reportProgressDone(c *conn.Conn, taskID uint64, success bool, errorMsg string) {
	c.WriteMessage(proto.MessageType_TASK_PROGRESS, &proto.TaskProgress{
		TaskId:  taskID,
		Done:    true,
		Success: success,
		Error:   errorMsg,
	})
}

// progressWriter 包装 io.Writer，每次 Write 时回调进度函数
type progressWriter struct {
	w          io.Writer
	total      int64
	written    int64
	progressFn func(done, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.written += int64(n)
	if pw.progressFn != nil && pw.total > 0 {
		pw.progressFn(pw.written, pw.total)
	}
	return n, err
}

var downloadClient = &http.Client{Timeout: 5 * time.Minute}

func downloadBinary(url, dest string, progressFn func(done, total int64)) error {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	var dst io.Writer = out
	if progressFn != nil && resp.ContentLength > 0 {
		dst = &progressWriter{w: out, total: resp.ContentLength, progressFn: progressFn}
	}

	_, err = io.Copy(dst, resp.Body)
	return err
}

func copyAndReplace(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
