package response

// 用户信息
type UserResponse struct {
	ID       uint64 `json:"id"`       // 用户ID
	Nickname string `json:"nickname"` // 用户昵称
	Avatar   string `json:"avatar"`   // 用户头像

	WorkspaceList []*WorkspaceResponse `json:"workspace_list"` // 工作空间列表
}
