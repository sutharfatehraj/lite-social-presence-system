package server

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
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func StartServer() {
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
	mgDAO := mongodao.InitMongoDao(client, db)

	// initialize the game server
	gamerServer, err := common.NewGameServer(mgDAO)
	if err != nil {
		fmt.Println("Error initializing game server", err)
		return
	}

	// initialize the user server
	userServer := common.NewUserServer()

	// keep checking game party duration in the background
	go func() {
		for {
			CheckPartyDuration(gamerServer, mgDAO)
			time.Sleep(1 * time.Minute)
		}
	}()

	// init services
	router.InitServices(mgDAO, userServer, gamerServer)

	fmt.Println("Starting the server...")

	// concurrently start REST API server and gRPC server
	go func() {
		r := router.InitRoutes()
		server := &http.Server{
			Handler: r,
			Addr:    literals.RestAPIServerAddress,
			// WriteTimeout: 200 * time.Second,
			// ReadTimeout:  200 * time.Second,
		}
		// start REST API server
		log.Fatal(server.ListenAndServe())
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// start gRPC server
		lis, err := net.Listen(literals.GRPCNetwork, literals.GRPCServerAddress)

		if err != nil {
			log.Fatalf("failed to listen: %v\n", err)
		}

		// create a gRPC server
		grpcServer := apis.InitStreamService(userServer, gamerServer)

		fmt.Printf("gRPC server started on %v address\n", literals.GRPCServerAddress)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v\n", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

// terminate the game party if the time is over
func CheckPartyDuration(gameServer *models.GameServer, mgDAO mongodao.MongoDAO) {

	var usersStatusToBeUpdated []string
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
			usersStatusToBeUpdated = append(usersStatusToBeUpdated, string(gameParty.CreatedBy))
			for userId, status := range gameParty.Players {
				if status == models.PlayerJoinedStatus {
					usersStatusToBeUpdated = append(usersStatusToBeUpdated, userId)
				}

			}
			partyIdsToBeTerminated = append(partyIdsToBeTerminated, partyId)
			delete(gameServer.Parties, partyId)
		}
	}
	gameServer.Mutex.Unlock()

	/*
		This approach is not efficient for larger dataset.
		For larger data set, terminate every party separately.
	*/
	if len(partyIdsToBeTerminated) > 0 {
		mgDAO.UpdateGamePartyStatus(context.TODO(), partyIdsToBeTerminated, models.GamePartyStatusOver)

		// update users status to idle
		mgDAO.UpdateUsersStatus(context.TODO(), usersStatusToBeUpdated, models.UserStatusIdle)
	}
}
