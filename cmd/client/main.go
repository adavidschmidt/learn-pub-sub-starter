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
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "army_moves."+usr, "army_moves.*", pubsub.SimpleQueueTransient, handlerMove(gameState, publishCh))
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "war", "war.*", pubsub.SimpleQueueDurable, handlerWar(gameState, publishCh))
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
