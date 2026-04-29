package errcode

const (
	// 消息相关错误
	MESSAGE_ERROR = 4000 + iota
	SmsSendCode
	SmsVerifyCode
)

var (
	ErrSmsVerifyCode = &AppError{Code: SmsVerifyCode, Msg: "验证码错误"}
	ErrSmsSendCode   = &AppError{Code: SmsSendCode, Msg: "发送验证码失败"}
)
