package errcode

const (
	// 客户端相关错误 3000 - 3999
	CLIENT_ERROR = 3000 + iota
	CLIENT_AUTH_FAILED
	CLIENT_NOT_READY
	CLIENT_DISCONNECT
	CLIENT_NOT_FOUND
)

var (
	ErrClientAuthFailed = &AppError{Code: CLIENT_AUTH_FAILED, Msg: "认证失败"}
	ErrClientNotReady   = &AppError{Code: CLIENT_NOT_READY, Msg: "客户端未就绪"}
	ErrClientDisconnect = &AppError{Code: CLIENT_DISCONNECT, Msg: "客户端断开"}
	ErrClientNotFound   = &AppError{Code: CLIENT_NOT_FOUND, Msg: "客户端未找到"}
)
