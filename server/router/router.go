package router

import (
	"fmt"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"lite-social-presence-system/server/apis"
	"net/http"

	"github.com/gorilla/mux"
)

// Fix: return r *mux.Router
func InitRoutes() *mux.Router {
	r := mux.NewRouter()

	// handlers

	// user APIs
	r.HandleFunc("/user/login", apis.UserLogInHandler).Methods(http.MethodPatch)
	r.HandleFunc("/user/logout", apis.UserLogOutHandler).Methods(http.MethodPatch)

	// friends APIs
	r.HandleFunc("/game/friends", apis.GetFriendsHandler).Methods(http.MethodGet) // /game/friends{id} -> then fetch using mux.Vars to getch path varaibles
	r.HandleFunc("/game/friends/request", apis.SendFriendRequestHandler).Methods(http.MethodPatch)
	r.HandleFunc("/game/friends/handle-request", apis.HandleFriendRequest).Methods(http.MethodPatch)
	r.HandleFunc("/game/friends/remove", apis.RemoveFriends).Methods(http.MethodDelete)

	// party APIs
	r.HandleFunc("/game/party/create", apis.CreateGameParty).Methods(http.MethodPost)
	r.HandleFunc("/game/party/invite", apis.InviteToGamePartyHandler).Methods(http.MethodPatch)
	r.HandleFunc("/game/party/handle", apis.HandleGamePartyInviteHandler).Methods(http.MethodPatch)
	r.HandleFunc("/game/party/join", apis.JoinGamePartyHandler).Methods(http.MethodPatch)
	r.HandleFunc("/game/party/exit", apis.ExitGamePartyHandler).Methods(http.MethodPatch)
	r.HandleFunc("/game/party/remove", apis.RemoveFromGamePartyHandler).Methods(http.MethodPatch)

	fmt.Printf("REST API server started on %v address\n", literals.RestAPIServerAddress)
	return r
}

// init services
func InitServices(mgDAO mongodao.MongoDAO, userServer *models.UserServer, gamerServer *models.GameServer) {

	// user services
	apis.InitUserLoginService(mgDAO, userServer)
	apis.InitUserLogOutService(mgDAO)

	// friends services
	apis.InitGetUsersService(mgDAO)
	apis.InitGetFriendsService(mgDAO)
	apis.InitSendFriendRequestService(mgDAO)
	apis.InitHandleFriendRequestService(mgDAO)
	apis.InitRemoveFriendsService(mgDAO)

	// game party services
	apis.InitCreateGamePartyService(gamerServer, mgDAO)
	apis.InitInviteToGamePartyService(gamerServer, mgDAO)
	apis.InitHandleGamePartyInviteService(gamerServer, mgDAO)
	apis.InitJoinGamePartyService(gamerServer, mgDAO)
	apis.InitExitGamePartyService(gamerServer, mgDAO)
	apis.InitRemoveUsersFromGamePartyService(gamerServer, mgDAO)
}
