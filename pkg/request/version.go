package request

// VersionUploadRequest 版本上传请求
type VersionUploadRequest struct {
	Version     uint32 `form:"version" binding:"required"`
	VersionName string `form:"version_name" binding:"required"`
	Os          string `form:"os" binding:"required"`
	Arch        string `form:"arch" binding:"required"`
	Changelog   string `form:"changelog"`
}
