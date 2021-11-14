package main

import (
	"encoding/json"
	"flag"
	"go-chat/messager"
	"go-chat/persistence"
	"go-chat/pkg"
	"go-chat/settings"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	pkg.RoomsMessages = make(map[string][]string)

	cfg := settings.GetConfig()

	// MySQL Connection
	db := persistence.Init(cfg)
	defer db.Close()

	// RabbitMQ connection
	amqp, err := messager.Connect(cfg)
	if err != nil {
		return
	}
	defer amqp.Close()
	ch, err := messager.OpenChannel()
	if err != nil {
		return
	}
	defer ch.Close()

	s := pkg.NewServer()
	go s.Run()

	msgs := messager.ReceiveMessageDeliveryChannel()

	go func() {
		for d := range msgs {
			var response messager.ClientMessage
			json.Unmarshal(d.Body, &response)

			hubs := *s.GetHubs()
			hub := hubs[response.HubName]
			hub.SendTo(response.Message, response.ClientRemoteAddress)
		}
	}()

	router := mux.NewRouter()
	pkg.SetRouterHandlerFuncs(router, s)

	log.Printf(" [*] Service started. To exit press CTRL+C")

	addr := flag.String("addr", ":8080", "http service address")
	// start server listen
	// with error handling
	log.Fatal(http.ListenAndServe(*addr, router))
}
