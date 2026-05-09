package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	downloadBinary(payload.DownloadUrl, tmpPath, func(done, total int64) {
		if total > 0 {
			pct := int32(10 + (done * 60 / total))
			if pct > 70 {
				pct = 70
			}
			reportProgress(c, task.TaskId, pct, fmt.Sprintf("下载中 %d%%", pct))
		}
	})

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

func downloadBinary(url, dest string, progressFn func(done, total int64)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if progressFn != nil {
		// 带进度回调的拷贝
		buf := make([]byte, 32*1024)
		var written int64
		total := resp.ContentLength
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					return werr
				}
				written += int64(n)
				progressFn(written, total)
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		return nil
	}

	_, err = io.Copy(out, resp.Body)
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
