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
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, "war", "war.*", pubsub.SimpleQueueDurable, handlerWar(gameState))
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

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		var acktype pubsub.Acktype
		switch moveOutcome {
		case gamelogic.MoveOutcomeSafe:
			acktype = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername()),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			acktype = pubsub.NackRequeue
		default:
			acktype = pubsub.NackDiscard
		}
		return acktype
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		warOutcome, _, _ := gs.HandleWar(rw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}
