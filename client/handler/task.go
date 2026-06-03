package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"time"

	pb "google.golang.org/protobuf/proto"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
	"github.com/rs/zerolog/log"
)

// TaskHandlerFunc 处理一种任务类型的函数签名。
type TaskHandlerFunc func(c *conn.Conn, task *proto.Task) error

// TaskHandler 分发服务端下发的任务到已注册的 TaskHandlerFunc。
// 通过 RegisterTask 注册新的任务类型，无需修改 Handle 方法。
type TaskHandler struct {
	onUpgradeSuccess func()
	taskHandlers     map[proto.TaskType]TaskHandlerFunc
}

// NewTaskHandler 创建一个任务处理器。返回具体类型以便调用 RegisterTask。
func NewTaskHandler(onUpgradeSuccess func()) *TaskHandler {
	h := &TaskHandler{
		onUpgradeSuccess: onUpgradeSuccess,
		taskHandlers:     make(map[proto.TaskType]TaskHandlerFunc),
	}
	h.RegisterTask(proto.TaskType_UPGRADE, h.handleUpgrade)
	h.RegisterTask(proto.TaskType_BATCH_COMMAND, h.handleBatchCommand)
	return h
}

// RegisterTask 注册一种任务类型的处理函数。
func (h *TaskHandler) RegisterTask(tt proto.TaskType, fn TaskHandlerFunc) {
	h.taskHandlers[tt] = fn
}

func (h *TaskHandler) Type() proto.MessageType {
	return proto.MessageType_TASK
}

func (h *TaskHandler) Handle(ctx conn.Context) error {
	var task proto.Task
	if err := ctx.Unmarshal(&task); err != nil {
		return err
	}

	log.Info().Int32("task_type", int32(task.TaskType)).Uint64("task_id", task.TaskId).Msg("收到任务")

	fn, ok := h.taskHandlers[task.TaskType]
	if !ok {
		log.Warn().Int32("task_type", int32(task.TaskType)).Msg("未注册的任务类型, 忽略")
		return nil
	}
	return fn(ctx.GetConn(), &task)
}

// BatchCommandPayload 批量命令任务的载荷
type BatchCommandPayload struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"` // 超时秒数，0 表示默认 30s
}

// BatchCommandResult 批量命令任务的执行结果
type BatchCommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

func (h *TaskHandler) handleBatchCommand(c *conn.Conn, task *proto.Task) error {
	var payload BatchCommandPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		return err
	}
	timeout := payload.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", payload.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := BatchCommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = err.Error()
		}
	}
	resultBytes, _ := json.Marshal(result)
	log.Info().Str("command", payload.Command).Int("exit_code", result.ExitCode).Msg("batch command executed")

	// 上报结果
	reportProgressDone(c, task.TaskId, result.ExitCode == 0, string(resultBytes))
	return nil
}

func (h *TaskHandler) handleUpgrade(c *conn.Conn, task *proto.Task) error {
	var payload proto.UpgradePayload
	if err := pb.Unmarshal(task.Payload, &payload); err != nil {
		return err
	}
	if err := executeUpgrade(c, task, &payload); err != nil {
		return err
	}
	if h.onUpgradeSuccess != nil {
		h.onUpgradeSuccess()
	}
	return nil
}
