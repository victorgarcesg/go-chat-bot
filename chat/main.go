// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-chat/messager"
	"go-chat/persistence"
	"go-chat/pkg"
	"go-chat/settings"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

func main() {
	var cfg settings.Config
	readFile(&cfg)

	// MySql Connection
	db := persistence.Init(&cfg)
	defer db.Close()

	// RabbitMQ connection
	amqp, err := messager.Connect(&cfg)
	if err != nil {
		return
	}
	defer amqp.Close()
	ch, err := messager.OpenChannel()
	if err != nil {
		return
	}
	defer ch.Close()

	flag.Parse()

	s := pkg.NewServer()
	go s.Run()

	msgs := messager.ReceiveMessageDeliveryChannel()

	go func() {
		for d := range msgs {
			var response messager.ClientMessage
			json.Unmarshal(d.Body, &response)

			fmt.Println("Message received: ")
			fmt.Println(response)
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

func readFile(cfg *settings.Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}
