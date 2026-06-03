package mq

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/rs/zerolog/log"
)

var pubSub *gochannel.GoChannel

func Init() error {
	pubSub = gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)
	return nil
}

func GetPubSub() *gochannel.GoChannel {
	if pubSub == nil {
		log.Fatal().Msg("pubsub not initialized")
	}
	return pubSub
}

func Close() error {
	if pubSub != nil {
		return pubSub.Close()
	}
	return nil
}
