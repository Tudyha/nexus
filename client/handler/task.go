package handler

import (
	pb "google.golang.org/protobuf/proto"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
)

type TaskHandler struct {
	onUpgradeSuccess func()
}

func NewTaskHandler(onUpgradeSuccess func()) conn.MessageHandler {
	return &TaskHandler{onUpgradeSuccess: onUpgradeSuccess}
}

func (h *TaskHandler) Type() proto.MessageType {
	return proto.MessageType_TASK
}

func (h *TaskHandler) Handle(ctx conn.Context) error {
	var task proto.Task
	if err := ctx.Unmarshal(&task); err != nil {
		return err
	}

	switch task.TaskType {
	case proto.TaskType_UPGRADE:
		var payload proto.UpgradePayload
		if err := pb.Unmarshal(task.Payload, &payload); err != nil {
			return err
		}
		if err := executeUpgrade(ctx.GetConn(), &task, &payload); err != nil {
			return err
		}
		// 升级成功，通知主循环重启
		if h.onUpgradeSuccess != nil {
			h.onUpgradeSuccess()
		}
		return nil
	case proto.TaskType_BATCH_COMMAND:
		// 预留
		return nil
	}
	return nil
}
