package response

import "time"

// 应用信息
type AppResponse struct {
	ID          uint64     `json:"id"`          // 应用ID
	Name        string     `json:"name"`        // 应用名称
	Description string     `json:"description"` // 应用描述
	Status      int        `json:"status"`      // 应用状态
	CreatedAt   *time.Time `json:"created_at"`  // 创建时间
}
