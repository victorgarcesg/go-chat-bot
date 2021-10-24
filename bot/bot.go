package main

import (
	"bot/messaging"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
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
			var response messaging.ClientMessage
			json.Unmarshal(d.Body, &response)

			data, err := readCSVFromUrl(messaging.STOCK_URL + response.Message)
			if err != nil {
				log.Fatal("Error parsing CSV from URL")
				return
			}

			dataFieldRows := data[1]
			stooqResponse := &messaging.StooqResponse{
				Symbol: dataFieldRows[0],
				Close:  dataFieldRows[6],
			}

			var message string
			if stooqResponse.Close != "N/A" {
				message = fmt.Sprintf("%s quote is %v per share.", stooqResponse.Symbol, stooqResponse.Close)
			} else {
				message = "Could not get stock quote."
			}

			clientMessage := &messaging.ClientMessage{HubName: response.HubName, ClientRemoteAddress: response.ClientRemoteAddress, Message: message}
			messaging.SendMessage(clientMessage)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rows [][]string
	for _, e := range data {
		rows = append(rows, strings.Split(e[0], ","))
	}

	return rows, nil
}
