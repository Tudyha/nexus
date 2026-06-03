package response

// 用户信息
type UserResponse struct {
	ID       uint64 `json:"id"`       // 用户ID
	Nickname string `json:"nickname"` // 用户昵称
	Avatar   string `json:"avatar"`   // 用户头像

	WorkspaceList []*WorkspaceResponse `json:"workspace_list"` // 工作空间列表
}

// UserItem 用户列表项
type UserItem struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}
