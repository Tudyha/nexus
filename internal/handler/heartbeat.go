package handler

import (
	"github.com/Tudyha/nexus/internal/model"
	"github.com/Tudyha/nexus/internal/service"
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/enum"
	"github.com/Tudyha/nexus/pkg/proto"
)

type HeartbeatHandler struct {
	clientService service.ClientService
}

func NewHeartbeatHandler() conn.MessageHandler {
	return &HeartbeatHandler{
		clientService: service.GetClientService(),
	}
}

func (h *HeartbeatHandler) Handle(ctx conn.Context) error {
	var req proto.HeartbeatReq
	if err := ctx.Unmarshal(&req); err != nil {
		return err
	}
	sessionId := getSessionId(ctx)

	client, err := h.clientService.GetBySessionID(ctx, sessionId)
	if err != nil {
		return err
	}

	clientId := client.ID

	if err := h.clientService.UpdateStatus(ctx, clientId, enum.ClientOnline); err != nil {
		return err
	}

	clientStat := &model.ClientStat{
		ClientID: clientId,

		CpuPercent: req.CpuStat.Percent,

		MemTotal:       req.MemStat.Total,
		MemUsed:        req.MemStat.Used,
		MemFree:        req.MemStat.Free,
		MemCached:      req.MemStat.Cached,
		MemBuffers:     req.MemStat.Buffers,
		MemAvailable:   req.MemStat.Available,
		MemUsedPercent: req.MemStat.UsedPercent,
		MemSwapTotal:   req.MemStat.SwapTotal,
		MemSwapUsed:    req.MemStat.SwapUsed,
		MemSwapFree:    req.MemStat.SwapFree,

		NetBytesSent: req.NetStat.BytesSent,
		NetBytesRecv: req.NetStat.BytesRecv,
	}

	if err := h.clientService.CreateClientStat(ctx, clientStat); err != nil {
		return err
	}

	// 回复 ACK，让客户端感知服务端在线
	return ctx.GetConn().WriteMessage(proto.MessageType_HEARTBEAT_ACK, &proto.Response{})
}

func (h *HeartbeatHandler) Type() proto.MessageType {
	return proto.MessageType_HEARTBEAT
}
