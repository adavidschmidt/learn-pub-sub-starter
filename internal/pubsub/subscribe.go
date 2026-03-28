package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

type Acktype int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queuetype SimpleQueueType,
	handler func(T) Acktype,
) error {
	chn, queue, err := DeclareAndBind(conn, exchange, queueName, key, queuetype)
	if err != nil {
		return err
	}
	msgs, err := chn.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			var val T
			err = json.Unmarshal(msg.Body, &val)
			if err != nil {
				fmt.Printf("Could not unmarshal message: %v\n", err)
				continue
			}
			acktype := handler(val)
			switch acktype {
			case Ack:
				fmt.Println("Ack")
				msg.Ack(false)
			case NackDiscard:
				fmt.Println("NackDiscard")
				msg.Nack(false, false)
			case NackRequeue:
				fmt.Println("NackRequeue")
				msg.Nack(false, true)
			}
		}
	}()
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	queue, err := chn.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable,
		queueType == SimpleQueueTransient,
		queueType == SimpleQueueTransient,
		false,
		amqp.Table{"x-dead-letter-exchange": "peril_dlx"},
	)
	if err != nil {
		return chn, amqp.Queue{}, err
	}
	err = chn.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return chn, queue, err
	}
	return chn, queue, nil
}
