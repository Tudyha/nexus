package errcode

const (
	APP_ERROR = 6000 + iota
	APP_NOT_FOUND
)

var (
	ErrAppNotFound = &AppError{Code: APP_NOT_FOUND, Msg: "找不到应用"}
)
