package enum

type LoginType uint

const (
	LoginTypeUnknown  LoginType = iota // 未知登录方式
	LoginTypeCode                      // 验证码登录
	LoginTypePassword                  // 密码登录
)
