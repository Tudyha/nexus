package errcode

const (
	TASK_ERROR      = 8000 + iota
	TASK_NOT_FOUND
)

var (
	ErrTaskNotFound = &AppError{Code: TASK_NOT_FOUND, Msg: "任务不存在"}
)
