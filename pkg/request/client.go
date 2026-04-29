package request

import "github.com/Tudyha/nexus/pkg/enum"

type ClientQueryRequest struct {
	PageQuery

	Status   enum.ClientStatus `json:"status" form:"status"`
	Hostname string            `json:"hostname" form:"hostname"`
}

type GenerateV2raySubscribeRequest struct {
	Ids []uint64 `form:"ids" json:"ids" binding:"required"`
}
