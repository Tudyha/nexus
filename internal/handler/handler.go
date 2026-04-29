package handler

import (
	"github.com/Tudyha/nexus/pkg/conn"
	constant "github.com/Tudyha/nexus/pkg/const"
)

// 获取会话id
func getSessionId(ctx conn.Context) string {
	return ctx.Value(constant.ContextKeySessionId).(string)
}
