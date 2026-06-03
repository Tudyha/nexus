package response

import "time"

// 工作空间信息
type WorkspaceResponse struct {
	ID          uint64     `json:"id"`          // 工作空间ID
	Name        string     `json:"name"`        // 工作空间名称
	Description string     `json:"description"` // 工作空间描述
	Status      int        `json:"status"`      // 工作空间状态
	CreatedAt   *time.Time `json:"created_at"`  // 创建时间

	AppList []*AppResponse `json:"app_list"` // 应用列表
}


// 工作空间用户
type WorkspaceUserResponse struct {
	ID          uint64    `json:"id"`
	WorkspaceID uint64    `json:"workspace_id"`
	UserID      uint64    `json:"user_id"`
	Role        int       `json:"role"`
	Nickname    string    `json:"nickname"`
	Avatar      string    `json:"avatar"`
	CreatedAt   time.Time `json:"created_at"`
}
