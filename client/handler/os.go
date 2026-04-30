package handler

import (
	"os"

	"github.com/Tudyha/nexus/pkg/conn"
	"github.com/Tudyha/nexus/pkg/proto"
)

type ExitHandler struct {
}

func NewExitHandler() conn.MessageHandler {
	return &ExitHandler{}
}

func (e *ExitHandler) Handle(ctx conn.Context) error {
	os.Exit(0)
	return nil
}

func (e *ExitHandler) Type() proto.MessageType {
	return proto.MessageType_EXIT
}
