package services

import (
	"context"
	"encoding/json"
	"hermes/cmd/config"
	"hermes/internals/api/presenter"

	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(message presenter.Message) error
}

type Relayer struct {
	connection *amqp.Connection
	config     *config.RabbitMqConfiguration
}

func NewRelayer(config *config.RabbitMqConfiguration) *Relayer {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil
	}
	return &Relayer{
		connection: conn,
		config:     config,
	}
}

func (r *Relayer) Publish(message presenter.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(r.config.Queue, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// convert message to json
	pmessage, _ := json.Marshal(message)

	ch.PublishWithContext(ctx, "", r.config.Queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        pmessage,
	})

	return nil
}

func (r *Relayer) Close() {
	r.connection.Close()
}