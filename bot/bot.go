package main

import (
	"bot/core"
	"bot/messaging"
	"encoding/json"
	"log"
)

func main() {
	amqp, err := messaging.Connect()
	if err != nil {
		return
	}
	defer amqp.Close()
	ch, err := messaging.OpenChannel()
	if err != nil {
		return
	}
	defer ch.Close()

	msgs := messaging.ReceiveMessageDeliveryChannel()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var cm messaging.ClientMessage
			json.Unmarshal(d.Body, &cm)
			message, err := core.GetStockQuote(cm.Message)
			response := messaging.ClientMessage{HubName: cm.HubName, ClientRemoteAddress: cm.ClientRemoteAddress, Message: message}
			if err != nil {
				log.Fatal(err)
				return
			}

			clientMessage := &messaging.ClientMessage{HubName: response.HubName, ClientRemoteAddress: response.ClientRemoteAddress, Message: response.Message}
			messaging.SendMessage(clientMessage)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
