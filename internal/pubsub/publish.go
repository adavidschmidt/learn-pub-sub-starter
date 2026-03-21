package pubsub

import (
	"context"
	"encoding/json"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

var simpleQueueTypeNames = map[SimpleQueueType]string{
	Durable:   "Durable",
	Transient: "Transient",
}

func ParseQueueType(str string) (SimpleQueueType, error) {
	for k, v := range simpleQueueTypeNames {
		if v == str {
			return k, nil
		}
	}
	return SimpleQueueType(-1), errors.New("invalid queue type: " + str)
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msg,
	})
	if err != nil {
		return err
	}
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType string,
) (*amqp.Channel, amqp.Queue, error) {
	chn, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}
	queue, err := chn.QueueDeclare(
		queueName,
		queueType == "durable",
		queueType == "transient",
		queueType == "transient",
		false,
		nil,
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
