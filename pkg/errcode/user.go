package errcode

const (
	USER_ERROR = 5000 + iota
	USER_DISABLED
	USER_NOT_EXIST
)

var (
	ErrUserDisabled = &AppError{Code: USER_DISABLED, Msg: "用户被禁用"}
	ErrUserNotExist = &AppError{Code: USER_NOT_EXIST, Msg: "用户不存在"}
)
