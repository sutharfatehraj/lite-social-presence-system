package main

import (
	"lite-social-presence-system/server/apis"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func InitRoutes() {
	r := mux.NewRouter()

	// handlers
	r.HandleFunc("/game/friends", apis.GetUsers).Methods("GET") // GetFriends

	server := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8081",
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
