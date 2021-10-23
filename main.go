// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"go-chat/articles"
	"go-chat/persistence"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var ServeHome = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))

		log.Println(r.URL)
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "./app/home.html")
	})

func main() {
	db := persistence.Init()
	defer db.Close()

	flag.Parse()

	s := articles.NewServer()
	go s.Run()

	router := mux.NewRouter()
	router.HandleFunc("/api/account/signup", persistence.UserSignup)
	router.HandleFunc("/api/account/login", persistence.UserLogin)
	router.HandleFunc("/", ServeHome).Methods("GET")
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
