package model

import "github.com/Tudyha/nexus/pkg/enum"

type Task struct {
	BaseModel
	TaskType int32 `gorm:"column:task_type"`
}

func (Task) TableName() string { return "t_task" }

type TaskExecution struct {
	BaseModel
	TaskID   uint64          `gorm:"column:task_id;index"`
	ClientID uint64          `gorm:"column:client_id;index"`
	TaskType int32           `gorm:"column:task_type"`
	Status   enum.TaskStatus `gorm:"column:status;default:0"`
	Progress int32           `gorm:"column:progress;default:0"`
	Message  string          `gorm:"column:message;type:text"`
	Error    string          `gorm:"column:error;type:text"`
}

func (TaskExecution) TableName() string { return "t_task_execution" }
