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
	// friends APIs
	r.HandleFunc("/game/friends", apis.GetFriends).Methods(http.MethodGet) // /game/friends{id} -> then fetch using mux.Vars to getch path varaibles
	r.HandleFunc("/game/friends/request", apis.SendFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/friends/handle-request", apis.HandleFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/friends/remove", apis.RemoveFriends).Methods(http.MethodDelete)

	// party APIs
	r.HandleFunc("/game/party/create", apis.CreateGameParty).Methods(http.MethodPost)
	r.HandleFunc("/game/party/invite", apis.InviteToGamePartyHandler).Methods(http.MethodPatch)
	return r
}
