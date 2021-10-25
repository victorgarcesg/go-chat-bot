package messager

import (
	"encoding/json"
	"fmt"
	"go-chat/persistence"
	"go-chat/settings"

	"github.com/streadway/amqp"
)

type ClientMessage struct {
	HubName             string `json:"hubName"`
	ClientRemoteAddress string `json:"clientRemoteAddress"`
	Message             string `json:"message"`
}

var Conn *amqp.Connection
var Channel *amqp.Channel
var ClientQueueName, StooqQueueName string

func Connect(cfg *settings.Config) (*amqp.Connection, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Pass,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port))

	if err != nil {
		persistence.FailOnError(err, "Failed to connect to RabbitMQ")
		return nil, err
	}

	Conn = conn

	ClientQueueName = cfg.RabbitMQ.ClientQueue
	StooqQueueName = cfg.RabbitMQ.StooqQueue

	return conn, nil
}

func OpenChannel() (*amqp.Channel, error) {
	ch, err := Conn.Channel()
	if err != nil {
		persistence.FailOnError(err, "Failed to open channel")
		return nil, err
	}

	Channel = ch

	return ch, nil
}

func SendMessage(message *ClientMessage) {
	ch, err := Conn.Channel()
	persistence.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		ClientQueueName, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
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
	q, err := Channel.QueueDeclare(
		StooqQueueName, // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	persistence.FailOnError(err, "Failed to declare a queue")

	msgs, err := Channel.Consume(
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
