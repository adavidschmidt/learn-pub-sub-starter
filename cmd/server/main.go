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
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Printf("Error connecting to rabbitMQ: %v", err)
		return
	}
	defer connection.Close()
	fmt.Println("Connection successful")
	chn, err := connection.Channel()
	if err != nil {
		log.Fatalf("Error creating channel: %v", err)
	}

	gamelogic.PrintServerHelp()
	for true {
		inputs := gamelogic.GetInput()
		if inputs == nil {
			continue
		}
		switch inputs[0] {
		case "pause":
			fmt.Println("Pausing")
			err = pubsub.PublishJSON(chn, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("UnPausing")
			err = pubsub.PublishJSON(chn, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Quitting")
			return
		default:
			fmt.Println("Uknown command")
		}
	}

}
