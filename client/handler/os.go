package handler

import (
	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
)

type ExitHandler struct {
	onExit func()
}

func NewExitHandler(onExit func()) conn.MessageHandler {
	return &ExitHandler{onExit: onExit}
}

func (e *ExitHandler) Handle(ctx conn.Context) error {
	if e.onExit != nil {
		e.onExit()
	}
	return nil
}

func (e *ExitHandler) Type() proto.MessageType {
	return proto.MessageType_EXIT
}
