package router

import (
	"lite-social-presence-system/server/apis"
	"net/http"

	"github.com/gorilla/mux"
)

// Fix: return r *mux.Router
func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	// handlers
	r.HandleFunc("/game/friends", apis.GetFriends).Methods(http.MethodGet)
	r.HandleFunc("/game/send-friend-request", apis.SendFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/handle-friend-request", apis.HandleFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/remove-freinds", apis.RemoveFriends).Methods(http.MethodDelete)

	return r
}
