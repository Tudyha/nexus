package errcode

const (
	// 版本相关错误 7000-7999
	VERSION_ERROR = 7000 + iota
	VERSION_NOT_FOUND
	VERSION_UPGRADE
	VERSION_DUPLICATE
	VERSION_NOT_READY
	LATEST_VERSION_NOT_FOUND
	VERSION_CREATE_FAIL
)

var (
	ErrVersionNotFound       = &AppError{Code: VERSION_NOT_FOUND, Msg: "版本不存在"}
	ErrVersionUpgrade        = &AppError{Code: VERSION_UPGRADE, Msg: "升级失败"}
	ErrVersionDuplicate      = &AppError{Code: VERSION_DUPLICATE, Msg: "版本号已存在"}
	ErrVersionNotReady       = &AppError{Code: VERSION_NOT_READY, Msg: "客户端不在线，无法升级"}
	ErrLatestVersionNotFound = &AppError{Code: LATEST_VERSION_NOT_FOUND, Msg: "已是最新版本"}
	ErrVersionCreateFail     = &AppError{Code: VERSION_CREATE_FAIL, Msg: "版本创建失败"}
)
