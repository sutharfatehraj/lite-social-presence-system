package main

import (
	"fmt"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/mongodao"
	"lite-social-presence-system/server/apis"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("Starting the API server...")

	// Get Client, Context, CalcelFunc and err from connect method.
	client, ctx, cancel, errConnecting := mongodao.Connect("mongodb://127.0.0.1:2717/") //"mongodb://localhost:27017")
	if errConnecting != nil {
		fmt.Println("Error connecting to mongoDB:", errConnecting)
		return
	}

	// Release resource when the main function is returned.
	defer mongodao.Close(client, ctx, cancel)
	// Ping mongoDB with Ping method
	mongodao.Ping(client, ctx)

	db := client.Database(literals.Database)
	mgDAO := mongodao.InitMongoDao(db)

	// initServices
	apis.InitGetUsersService(mgDAO)
	apis.InitGetFriendsService(mgDAO)
	apis.InitSendFriendRequestService(mgDAO)
	apis.InitHandleFriendRequestService(mgDAO)
	apis.InitRemoveFriendsService(mgDAO)

	r := mux.NewRouter()

	// handlers
	r.HandleFunc("/game/friends", apis.GetFriends).Methods(http.MethodGet)
	r.HandleFunc("/game/send-friend-request", apis.SendFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/handle-friend-request", apis.HandleFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/remove-freinds", apis.RemoveFriends).Methods(http.MethodDelete)

	server := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8081",
		// WriteTimeout: 200 * time.Second,
		// ReadTimeout:  200 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}
