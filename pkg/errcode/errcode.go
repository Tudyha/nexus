package errcode

const (
	INVALID_PARAMS        = 400
	UNAUTHORIZED          = 401
	FORBIDDEN             = 403
	NOT_FOUND             = 404
	INTERNAL_SERVER_ERROR = 500
)

var commondErrorMessages = map[int]string{

	INVALID_PARAMS:        "参数错误",
	UNAUTHORIZED:          "未授权",
	FORBIDDEN:             "无权限",
	NOT_FOUND:             "资源不存在",
	INTERNAL_SERVER_ERROR: "服务器内部错误",
}

var (
	// 通用错误
	ErrInvalidParams  = &AppError{Code: INVALID_PARAMS, Msg: commondErrorMessages[INVALID_PARAMS]}
	ErrUnauthorized   = &AppError{Code: UNAUTHORIZED, Msg: commondErrorMessages[UNAUTHORIZED]}
	ErrForbidden      = &AppError{Code: FORBIDDEN, Msg: commondErrorMessages[FORBIDDEN]}
	ErrNotFound       = &AppError{Code: NOT_FOUND, Msg: commondErrorMessages[NOT_FOUND]}
	ErrInternalServer = &AppError{Code: INTERNAL_SERVER_ERROR, Msg: commondErrorMessages[INTERNAL_SERVER_ERROR]}
)

// AppError 应用错误
type AppError struct {
	Code int
	Msg  string
}

func (e *AppError) Error() string {
	return e.Msg
}
