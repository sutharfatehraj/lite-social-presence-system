package apis

import (
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"lite-social-presence-system/protos/gampepb"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var userServiceOnce sync.Once

type userService struct {
	gampepb.UnimplementedUserServiceServer
	userServer *models.UserServer
	gameServer *models.GameServer
}

func InitStreamService(usrSrvr *models.UserServer, gamerSrvr *models.GameServer) *grpc.Server {
	// create a gRPC server
	grpcServer := grpc.NewServer()

	userServiceOnce.Do(func() {
		userServiceStruct := &userService{
			userServer: usrSrvr,
			gameServer: gamerSrvr,
		}
		gampepb.RegisterUserServiceServer(grpcServer, userServiceStruct)
	})

	return grpcServer
}

func (s userService) StreamUserStatusChange(requestData *gampepb.UserStatusChangeRequest, stream gampepb.UserService_StreamUserStatusChangeServer) error {

	log.Printf("stream friend online status update for userId : %v", requestData.UserId)

	var errMsg string

	if requestData.UserId != literals.EmptyString {
		var wg sync.WaitGroup

		// initialize the channel
		s.userServer.UserDetails[requestData.UserId] = &models.UserDetails{
			FriendOnlineUpdateMsg: make(chan string),
		}
		for msg := range s.userServer.UserDetails[requestData.UserId].FriendOnlineUpdateMsg {
			wg.Add(1)
			go func(msg string) {
				defer wg.Done()
				time.Sleep(1 * time.Second)
				resp := gampepb.UserStatusChangeResponse{
					Message: msg,
				}

				if err := stream.Send(&resp); err != nil {
					log.Printf("send error %v\n", err)
				}
				log.Printf("finishing sending the message : %v\n", msg)
			}(msg)
		}
		wg.Wait()
	} else {
		errMsg = "empty userId"
	}

	if errMsg != literals.EmptyString {
		log.Println(errMsg)
		return status.Errorf(codes.InvalidArgument, errMsg)
	}
	return nil
}

// will stream the message to the userId whenever a player joins the game
// userId can be of the one who created the party or who has joined the game
func (s userService) StreamPlayerJoinedStatus(requestData *gampepb.PlayerInPartyRequest, stream gampepb.UserService_StreamPlayerJoinedStatusServer) error {

	log.Printf("stream player joined message for userId : %v", requestData.PartyId)
	var errMsg string

	if s.gameServer.Parties != nil {
		if gameParty, ok := s.gameServer.Parties[requestData.PartyId]; ok {
			userFound := false

			for player, playerStatus := range gameParty.Players {
				if player == requestData.UserId && playerStatus == models.PlayerJoinedStatus {
					userFound = true
					break
				}
			}
			if userFound || gameParty.CreatedBy == requestData.UserId {

				var wg sync.WaitGroup

				// initialize the channel to listen to any player joining
				// gameParty.PlayerStatusUpdateMsg = make(chan string) // unbuffered channel
				gameParty.PlayerStatusUpdateMsg = make(chan string, 1) // channel with capacity 1 so that sender does not get blocked when sending data for 1 user and can proceed with its work

				for msg := range gameParty.PlayerStatusUpdateMsg {
					wg.Add(1)
					go func(msg string) {
						defer wg.Done()
						time.Sleep(1 * time.Second)
						resp := &gampepb.PlayersInPartyResponse{
							Message: msg,
						}
						if err := stream.Send(resp); err != nil {
							log.Printf("send error %v\n", err)
						}
						log.Printf("finishing sending the message : %v\n", msg)
					}(msg)
				}
				wg.Wait()
				// close(gameParty.PlayerStatusUpdateMsg) //close the channel here?
			} else {
				errMsg = "invalid userId. User has either not created the party or has not joined the party"
			}
		} else {
			errMsg = "party not found"
		}
	} else {
		errMsg = "no parties created"
	}

	if errMsg != literals.EmptyString {
		log.Println(errMsg)
		return status.Errorf(codes.InvalidArgument, errMsg)
	}

	return nil
}
