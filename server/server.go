package main

import (
	"context"
	"fmt"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"lite-social-presence-system/server/apis"
	"lite-social-presence-system/server/common"
	"lite-social-presence-system/server/router"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Get Client, Context, CalcelFunc and err from connect method.
	client, ctx, cancel, errConnecting := mongodao.Connect("mongodb://127.0.0.1:2717/") //"mongodb://localhost:2717")
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

	// initialize the game server
	gamerServer, err := common.NewGameServer(mgDAO)
	if err != nil {
		fmt.Println("Error initializing game server", err)
		return
	}

	// keep checking game party duration in the background
	// think of some better appraoch to check party duration
	go func() {
		for {
			CheckPartyDuration(gamerServer, mgDAO)
			time.Sleep(1 * time.Minute)
		}
	}()

	fmt.Println("Starting the API server...")

	// on joining user ID will be moved from accepted array, if present else return, to players array

	// initServices
	apis.InitGetUsersService(mgDAO)
	apis.InitGetFriendsService(mgDAO)
	apis.InitSendFriendRequestService(mgDAO)
	apis.InitHandleFriendRequestService(mgDAO)
	apis.InitRemoveFriendsService(mgDAO)
	apis.InitCreateGamePartyService(gamerServer, mgDAO)

	r := router.InitRoutes()
	server := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8081",
		// WriteTimeout: 200 * time.Second,
		// ReadTimeout:  200 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

// terminate the game party if the time is over
func CheckPartyDuration(gameServer *models.GameServer, mgDAO mongodao.MongoDAO) {
	var partyIdsToBeTerminated []string
	gameServer.Mutex.Lock()
	for partyId, gameParty := range gameServer.Parties {
		if time.Since(gameParty.StartTime) > gameParty.Duration {
			logrus.WithFields(logrus.Fields{
				literals.LLCurrentTimeInUTC:        time.Now().UTC().String(),
				literals.LLGamePartyStartTimeInUTC: gameParty.StartTime.String(),
				literals.LLGamePartyDuration:       gameParty.Duration,
				literals.LLPartyId:                 partyId,
			}).Info("Terminating party")
			partyIdsToBeTerminated = append(partyIdsToBeTerminated, partyId)
			delete(gameServer.Parties, partyId)
		}
	}
	gameServer.Mutex.Unlock()

	/*
		This approach is not to be used for larger dataset.
		For larger data set, terminate every party separately.
	*/
	if len(partyIdsToBeTerminated) > 0 {
		mgDAO.UpdateGamePartyStatus(context.TODO(), partyIdsToBeTerminated, models.GamePartyStatusOver)
	}
}
