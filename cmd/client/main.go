package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()
	usr, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create channel: %v", err)
	}
	gameState := gamelogic.NewGameState(usr)
	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, "pause."+usr, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gameState))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves."+usr, "army_moves.*", pubsub.SimpleQueueTransient, handlerMove(gameState))
	for {
		inputs := gamelogic.GetInput()
		if len(inputs) == 0 {
			continue
		}
		switch inputs[0] {
		case "spawn":
			err = gameState.CommandSpawn(inputs)
			if err != nil {
				fmt.Printf("Error spawning unit: %v\n", err)
			}
		case "move":
			move, err := gameState.CommandMove(inputs)
			if err != nil {
				fmt.Printf("Error moving unit: %v\n", err)
			}
			err = pubsub.PublishJSON(publishCh, routing.ExchangePerilTopic, "army_moves."+usr, move)
			if err != nil {
				fmt.Printf("Error moving unit: %v\n", err)
			}
			fmt.Printf("Moved unit %v to %v\n", move.Units, move.ToLocation)
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		var acktype pubsub.Acktype
		switch moveOutcome {
		case gamelogic.MoveOutcomeSafe:
			acktype = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			acktype = pubsub.Ack
		default:
			acktype = pubsub.NackDiscard
		}
		return acktype
	}
}
