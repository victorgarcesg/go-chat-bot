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
	setRouterHandlerFuncs(router, s)

	log.Printf(" [*] Service started. To exit press CTRL+C")

	addr := flag.String("addr", ":8080", "http service address")
	// start server listen
	// with error handling
	log.Fatal(http.ListenAndServe(*addr, router))
}

func setRouterHandlerFuncs(router *mux.Router, s *pkg.Server) {
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

func getListRooms(s *pkg.Server, response http.ResponseWriter, request *http.Request) {
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
