package model

import (
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/proto"
)

type Tunnel struct {
	BaseModel

	ClientID   uint64            `gorm:"column:client_id;not null;index"` // 客户端id
	TunnelType proto.TunnelType  `gorm:"column:tunnel_type;not null"`     // 隧道类型
	LocalPort  uint16            `gorm:"column:local_port;not null"`      // 本地端口
	RemoteAddr string            `gorm:"column:remote_addr;not null"`     // 远程地址
	Status     enum.TunnelStatus `gorm:"column:status;not null"`          // 隧道状态
}

func (Tunnel) TableName() string {
	return "t_tunnel"
}
