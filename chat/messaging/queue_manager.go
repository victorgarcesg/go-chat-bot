package messaging

import (
	"encoding/json"
	"fmt"
	"go-chat/persistence"

	"github.com/streadway/amqp"
)

type ClientMessage struct {
	HubName             string `json:"hubName"`
	ClientRemoteAddress string `json:"clientRemoteAddress"`
	Message             string `json:"message"`
}

const (
	CLIENT_QUEUE_NAME = "chat_bot_client"
	STOOQ_QUEUE_NAME  = "chat_bot_stooq"
)

var CONN *amqp.Connection
var CHANNEL *amqp.Channel

func Connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		persistence.FailOnError(err, "Failed to connect to RabbitMQ")
		return nil, err
	}

	CONN = conn

	return conn, nil
}

func OpenChannel() (*amqp.Channel, error) {
	ch, err := CONN.Channel()
	if err != nil {
		persistence.FailOnError(err, "Failed to open channel")
		return nil, err
	}

	CHANNEL = ch

	return ch, nil
}

func SendMessage(message *ClientMessage) {
	ch, err := CONN.Channel()
	persistence.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		CLIENT_QUEUE_NAME, // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	persistence.FailOnError(err, "Failed to declare a queue")

	json, err := json.Marshal(message)
	if err != nil {
		persistence.FailOnError(err, "Failed to parse body message")
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		})
	persistence.FailOnError(err, "Failed to publish a message")

	fmt.Printf("Message sent: %s\n", json)
}

func ReceiveMessageDeliveryChannel() <-chan amqp.Delivery {
	q, err := CHANNEL.QueueDeclare(
		STOOQ_QUEUE_NAME, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	persistence.FailOnError(err, "Failed to declare a queue")

	msgs, err := CHANNEL.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	persistence.FailOnError(err, "Failed to register a consumer")

	return msgs
}
