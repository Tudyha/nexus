package request

import "github.com/Tudyha/nexus/pkg/proto"

type TunnelCreateRequest struct {
	ClientID   uint             `json:"client_id" binding:"required"`
	TunnelType proto.TunnelType `json:"tunnel_type" binding:"required"`
	LocalPort  uint16           `json:"local_port" binding:"required"`
	RemoteAddr string           `json:"remote_addr" binding:"required"`
}
