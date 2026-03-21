package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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
	_, _, err = pubsub.DeclareAndBind(conn, "peril_direct", routing.PauseKey+"."+usr, routing.PauseKey, "transient")
	if err != nil {
		log.Fatalf("Error making and binding queue: %v", err)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Exiting...")
}
