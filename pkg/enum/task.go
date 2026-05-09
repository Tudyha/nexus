package enum

type TaskStatus int

const (
	TaskStatusPending   TaskStatus = iota // 待执行
	TaskStatusRunning                     // 执行中
	TaskStatusInterrupt                   // 异常中断
	TaskStatusSuccess                     // 执行成功
	TaskStatusFailed                      // 执行失败
)
