package apis

import (
	"fmt"
	"lite-social-presence-system/models"
	"lite-social-presence-system/protos/gampepb"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var userServiceOnce sync.Once

type userService struct {
	gampepb.UnimplementedUserServiceServer
	gameServer *models.GameServer
}

func InitStreamService(gamerServer *models.GameServer) *grpc.Server {
	// create a gRPC server
	grpcServer := grpc.NewServer()

	userServiceOnce.Do(func() {
		userServiceStruct := &userService{
			gameServer: gamerServer,
		}
		gampepb.RegisterUserServiceServer(grpcServer, userServiceStruct)
	})

	return grpcServer
}

func (s userService) StreamUserStatusChange(requestData *gampepb.UserStatusChangeRequest, stream gampepb.UserService_StreamUserStatusChangeServer) error {

	log.Printf("fetch response for userId : %v", requestData.UserId)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(count int64) {
			defer wg.Done()
			time.Sleep(time.Duration(count) * time.Second)
			resp := gampepb.UserStatusChangeResponse{
				FriendId: fmt.Sprint(i),
				Name:     "Fj" + fmt.Sprint(i),
				Message:  fmt.Sprint(requestData.UserId + "is online"),
			}

			if err := stream.Send(&resp); err != nil {
				log.Printf("send error %v\n", err)
			}
			log.Printf("finishing request number : %d", count)
		}(int64(i))
	}

	wg.Wait()
	return nil
}

// will stream the message to the userId whenever a player joins the game
func (s userService) StreamPlayerJoinedStatus(requestData *gampepb.PlayerInPartyRequest, stream gampepb.UserService_StreamPlayerJoinedStatusServer) error {

	log.Printf("fetch response for userId : %v", requestData.PartyId)

	var wg sync.WaitGroup
	if s.gameServer.Parties != nil {
		if gameParty, ok := s.gameServer.Parties[requestData.PartyId]; ok {

			// initialize the channel to listen to any player joining
			// gameParty.PlayerStatusUpdateMsg = make(chan string) // unbuffered channel
			gameParty.PlayerStatusUpdateMsg = make(chan string, 1) // channel with capacity 1 so that sender does not get blocked when sending data for 1 user and can proceed with its work

			for msg := range gameParty.PlayerStatusUpdateMsg {
				wg.Add(1)
				go func(msg string) {
					defer wg.Done()
					time.Sleep(1 * time.Second)
					resp := gampepb.PlayersInPartyResponse{
						Message: msg,
					}
					if err := stream.Send(&resp); err != nil {
						log.Printf("send error %v\n", err)
					}
					log.Printf("finishing sending the message : %v\n", msg)
				}(msg)
			}
			wg.Wait()
			// close(gameParty.PlayerStatusUpdateMsg) //close the channel here??
		} else {
			log.Println("party not found")
		}
	}

	return nil
}
