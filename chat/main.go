// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-chat/articles"
	"go-chat/messaging"
	"go-chat/persistence"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var serveHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/home" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/home.html")
	})

var chatRoom = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/chat.html")
	})

var loginForm = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/login.html")
	})

var signupForm = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/signup" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/signup.html")
	})

func main() {
	// MySql Connection
	db := persistence.Init()
	defer db.Close()

	// RabbitMQ connection
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

	flag.Parse()

	s := articles.NewServer()
	go s.Run()

	msgs := messaging.ReceiveMessageDeliveryChannel()

	go func() {
		for d := range msgs {
			var response messaging.ClientMessage
			json.Unmarshal(d.Body, &response)

			fmt.Println("Message received: ")
			fmt.Println(response)
			hubs := *s.GetHubs()
			hub := hubs[response.HubName]
			hub.SendTo(response.Message, response.ClientRemoteAddress)
		}
	}()

	router := mux.NewRouter()
	setRouterHandlerFuncs(router, s)

	log.Printf(" [*] Service started. To exit press CTRL+C")

	addr := flag.String("addr", ":8080", "http service address")
	// start server listen
	// with error handling
	log.Fatal(http.ListenAndServe(*addr, router))
}

func setRouterHandlerFuncs(router *mux.Router, s *articles.Server) {
	router.HandleFunc("/api/account/signup", persistence.UserSignup).Methods("POST")
	router.HandleFunc("/api/account/login", persistence.UserLogin).Methods("POST")
	router.HandleFunc("/api/rooms", func(rw http.ResponseWriter, r *http.Request) {
		getListRooms(s, rw, r)
	}).Methods("GET")
	router.HandleFunc("/home", serveHome)
	router.HandleFunc("/chat", chatRoom)
	router.HandleFunc("/signup", signupForm)
	router.HandleFunc("/", loginForm)
	router.HandleFunc("/ws", s.ServeWs)
}

func getListRooms(s *articles.Server, response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	var hubsNames []string
	hubs := s.GetHubs()
	for name, _ := range *hubs {
		hubsNames = append(hubsNames, name)
	}

	json, err := json.Marshal(hubsNames)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}

	response.Write(json)
}
