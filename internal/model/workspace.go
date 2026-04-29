package model

type Workspace struct {
	BaseModel

	Name        string `gorm:"column:name;type:varchar(64);not null;comment:工作空间名称"`
	Description string `gorm:"column:description;type:varchar(255);comment:工作空间描述"`
	Status      int    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态(0-禁用 1-启用)"`
}

func (Workspace) TableName() string {
	return "t_workspace"
}

type WorkspaceUser struct {
	BaseModel

	WorkspaceID uint64 `gorm:"column:workspace_id;type:bigint unsigned;not null;comment:工作空间id"`
	UserID      uint64 `gorm:"column:user_id;type:bigint unsigned;not null;comment:用户id"`
	Role        int    `gorm:"column:role;type:tinyint;not null;default:1;comment:角色(1-管理员 2-成员)"`
}

func (WorkspaceUser) TableName() string {
	return "t_workspace_user"
}
