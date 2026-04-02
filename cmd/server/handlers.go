package main

import (
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.Acktype {
	return func(msg routing.GameLog) pubsub.Acktype {
		err := gamelogic.WriteLog(msg)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
