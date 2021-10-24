package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type ClientMessage struct {
	HubName             string `json:"hubName"`
	ClientRemoteAddress string `json:"clientRemoteAddress"`
	Message             string `json:"message"`
}

const (
	STOOQ_QUEUE_NAME  = "chat_bot_stooq"
	CLIENT_QUEUE_NAME = "chat_bot_client"
	STOCK_URL         = "https://stooq.com/q/l/?f=sd2t2ohlcv&h&e=csv&s="
)

var CONN *amqp.Connection
var CHANNEL *amqp.Channel

func Connect() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		return nil, err
	}

	CONN = conn

	return conn, nil
}

func OpenChannel() (*amqp.Channel, error) {
	ch, err := CONN.Channel()
	if err != nil {
		failOnError(err, "Failed to open channel")
		return nil, err
	}

	CHANNEL = ch

	return ch, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func ReceiveMessageDeliveryChannel() <-chan amqp.Delivery {
	q, err := CHANNEL.QueueDeclare(
		CLIENT_QUEUE_NAME, // name
		false,             // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := CHANNEL.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	return msgs
}

func SendMessage(message *ClientMessage) {
	q, err := CHANNEL.QueueDeclare(
		STOOQ_QUEUE_NAME, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare a queue")

	json, err := json.Marshal(message)
	if err != nil {
		failOnError(err, "Failed to parse body message")
	}

	err = CHANNEL.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		})
	failOnError(err, "Failed to publish a message")

	fmt.Printf("Message sent: %s\n", json)
}
