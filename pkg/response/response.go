package response

import (
	"errors"
	"net/http"

	"github.com/Tudyha/nexus/pkg/errcode"
	"github.com/gin-gonic/gin"
)

// 响应结构体
type Response struct {
	Code int    `json:"code"` // 状态码, 0: 成功
	Msg  string `json:"msg"`  // 错误信息
	Data any    `json:"data"` // 数据
}

func Success(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "",
		Data: data,
	})
}

type Page[T any] struct {
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

func Fail(ctx *gin.Context, err error) {
	FailWithMsg(ctx, err, "")
}

func FailWithMsg(ctx *gin.Context, err error, msg string) {
	code, errMsg := getErrorMsg(err)
	if msg == "" {
		msg = errMsg
	}

	ctx.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

func getErrorMsg(err error) (code int, msg string) {
	var appErr *errcode.AppError
	if errors.As(err, &appErr) {
		if msg == "" {
			msg = appErr.Msg
		}
		code = appErr.Code
	} else {
		if msg == "" {
			msg = err.Error()
		}
		code = errcode.ErrInternalServer.Code
	}
	return
}
