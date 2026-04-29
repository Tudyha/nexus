package response

import (
	"time"

	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/proto"
)

type TunnelResponse struct {
	ID        uint64     `json:"id"`
	CreatedAt *time.Time `json:"created_at"`

	ClientID   uint64            `json:"client_id"`
	TunnelType proto.TunnelType  `json:"tunnel_type"`
	LocalPort  uint16            `json:"local_port"`
	RemoteAddr string            `json:"remote_addr"`
	Status     enum.TunnelStatus `json:"status"`
}
