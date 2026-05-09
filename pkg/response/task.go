package response

import (
	"time"

	"github.com/Tudyha/nexus/pkg/enum"
)

type TaskResponse struct {
	ID       uint64 `json:"id"`
	TaskType int32  `json:"task_type"`
	Status   int32  `json:"status"`
}

type TaskExecutionResponse struct {
	ID        uint64          `json:"id"`
	TaskID    uint64          `json:"task_id"`
	ClientID  uint64          `json:"client_id"`
	TaskType  int32           `json:"task_type"`
	Status    enum.TaskStatus `json:"status"`
	Progress  int32           `json:"progress"`
	Message   string          `json:"message"`
	Error     string          `json:"error"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
