package errcode

const (
	// 认证相关错误 2000-2999
	AUTH_ERROR = 2000 + iota
	AUTH_FAILED
	AUTH_NOT_SET_PASSWORD
)

var (
	// 认证相关错误
	ErrAuth               = &AppError{Code: AUTH_ERROR, Msg: "认证错误"}
	ErrLoginFailed        = &AppError{Code: AUTH_FAILED, Msg: "登录失败，用户名或密码错误"}
	ErrAuthNotSetPassword = &AppError{Code: AUTH_NOT_SET_PASSWORD, Msg: "请先设置密码"}
)
