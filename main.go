// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go-chat/articles"
	"go-chat/persistence"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/rs/cors"
)

var ServeHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/home" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, "./templates/home.html")
	})

var LoginForm = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			http.ServeFile(w, r, "./templates/login.html")
		} else {
			user := readForm(r)
			fmt.Fprintf(w, "Hello "+user.Username+"!")
		}
	})

func readForm(r *http.Request) *persistence.User {
	r.ParseForm()
	user := new(persistence.User)
	decoder := schema.NewDecoder()
	decodeErr := decoder.Decode(user, r.PostForm)
	if decodeErr != nil {
		log.Printf("error mapping parsed form data to struct : %s", decodeErr)
	}

	return user
}

func main() {
	db := persistence.Init()
	defer db.Close()

	flag.Parse()

	s := articles.NewServer()
	go s.Run()

	router := mux.NewRouter()
	router.HandleFunc("/api/account/signup", persistence.UserSignup).Methods("POST")
	router.HandleFunc("/api/account/login", persistence.UserLogin).Methods("POST")
	router.HandleFunc("/home", ServeHome)
	router.HandleFunc("/", LoginForm)
	router.HandleFunc("/ws", s.ServeWs)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowCredentials: true,
	})

	addr := flag.String("addr", ":8080", "http service address")
	handler := c.Handler(router)

	// start server listen
	// with error handling
	log.Fatal(http.ListenAndServe(*addr, handler))
}
