package model

type App struct {
	BaseModel

	WorkspaceID uint64 `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	AppSecret   string `gorm:"not null"` // HMAC 密钥
	Description string `json:"description"`
	Config      string `json:"config" gorm:"type:text"` // 客户端配置覆盖（JSON）
	Status      int    `json:"status"`
}

func (a *App) TableName() string {
	return "t_app"
}
