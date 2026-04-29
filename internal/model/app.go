package model

type App struct {
	BaseModel

	WorkspaceID uint64 `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	AppSecret   string `gorm:"not null"` // HMAC 密钥
	Description string `json:"description"`
	Status      int
}

func (a *App) TableName() string {
	return "t_app"
}
