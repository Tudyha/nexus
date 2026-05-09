package v1

import (
	"path/filepath"

	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/internal/session"
	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/Tudyha/nexus/pkg/request"
	"github.com/Tudyha/nexus/pkg/response"
	"github.com/Tudyha/nexus/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type VersionController struct {
	versionService service.VersionService
	taskService    service.TaskService
	sessionManager session.Manager
}

func newVersionController() *VersionController {
	return &VersionController{
		versionService: service.GetVersionService(),
		sessionManager: session.GetManager(),
		taskService:    service.GetTaskService(),
	}
}

// Upload 上传客户端二进制文件
func (h *VersionController) Upload(ctx *gin.Context) {
	var req request.VersionUploadRequest
	if err := ctx.ShouldBind(&req); err != nil {
		response.FailWithMsg(ctx, errcode.ErrInvalidParams, err.Error())
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	// 保存上传的文件
	dst := filepath.Join("./tmp/", uuid.NewString())
	ctx.SaveUploadedFile(file, dst)
	if err := h.versionService.Upload(ctx, req.Version, req.VersionName, req.Os, req.Arch, req.Changelog, dst, file.Filename); err != nil {
		response.FailWithMsg(ctx, errcode.ErrVersionCreateFail, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// Page 版本列表
func (h *VersionController) Page(ctx *gin.Context) {
	var req request.PageQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}
	res, err := h.versionService.GetPage(ctx, req)
	if err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, res)
}

// Delete 删除版本
func (h *VersionController) Delete(ctx *gin.Context) {
	id := utils.StringToUint64(ctx.Param("id"))
	if id == 0 {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	if err := h.versionService.Delete(ctx, id); err != nil {
		response.Fail(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// Latest 获取最新版本
func (h *VersionController) Latest(ctx *gin.Context) {
	goos := ctx.Query("os")
	arch := ctx.Query("arch")
	if goos == "" || arch == "" {
		response.Fail(ctx, errcode.ErrInvalidParams)
		return
	}

	v, err := h.versionService.GetLatestByOS(ctx, goos, arch)
	if err != nil {
		response.Fail(ctx, errcode.ErrVersionNotFound)
		return
	}
	response.Success(ctx, response.VersionResponse{
		ID:          v.ID,
		Version:     v.Version,
		VersionName: v.VersionName,
		Os:          v.Os,
		Arch:        v.Arch,
		Checksum:    v.Checksum,
		BinarySize:  v.BinarySize,
		Changelog:   v.Changelog,
		FileName:    v.FileName,
	})
}
