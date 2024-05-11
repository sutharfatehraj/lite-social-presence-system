package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"net/http"
	"sync"
)

type JoinGamePartyService interface {
	ValidateRequest(ctx context.Context, requestData *models.JoinGamePartyRequestData) []string
	JoinGameParty(ctx context.Context, requestData *models.JoinGamePartyRequestData) error
}

var joinGamePartyServiceStruct JoinGamePartyService
var joinGamePartyServiceOnce sync.Once

type joinGamePartyService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitJoinGamePartyService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) JoinGamePartyService {
	joinGamePartyServiceOnce.Do(func() {
		joinGamePartyServiceStruct = &joinGamePartyService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return joinGamePartyServiceStruct
}

func GetJoinGamePartyService() JoinGamePartyService {
	if joinGamePartyServiceStruct == nil {
		panic("JoinGameParty Service not initialized")
	}
	return joinGamePartyServiceStruct
}

func (c joinGamePartyService) ValidateRequest(ctx context.Context, requestData *models.JoinGamePartyRequestData) []string {
	var errs []error
	var errorString []string

	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	}

	if requestData.PartyId == literals.EmptyString {
		errs = append(errs, errors.New("empty partyId in the request data"))
	}

	if errs == nil {
		// user should be present in party as a player and should have status as accepted
		if _, ok := c.gameServer.Parties[requestData.PartyId]; !ok {
			errs = append(errs, errors.New("invalid partyId in the request data"))
		} else if c.gameServer.Parties[requestData.PartyId].Players != nil {
			if playerStatus, ok := c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId]; ok {
				// player can join if his current status is "accepted"
				if playerStatus != models.PlayerAcceptedStatus {
					errs = append(errs, errors.New("player "+requestData.UserId+" has current status: "+string(playerStatus)+". cannot update decision to "+string(models.PlayerJoinedStatus)))
				}
			} else {
				errs = append(errs, errors.New("player not found in the game party"))
			}
		} else {
			errs = append(errs, errors.New("no players found in the game party"))
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			errorString = append(errorString, err.Error())
		}
		return errorString
	}

	return nil
}

func JoinGamePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.JoinGamePartyResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.JoinGamePartyRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read join game party message: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal join game party message : %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetJoinGamePartyService()

	errStrings = svc.ValidateRequest(ctx, requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.JoinGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to join the game party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c joinGamePartyService) JoinGameParty(ctx context.Context, requestData *models.JoinGamePartyRequestData) error {

	err := c.mongoDAO.UpdatePlayersDecisionForGameParty(ctx, requestData.PartyId, []string{requestData.UserId}, models.PlayerJoinedStatus)
	if err != nil {
		return err
	}

	_, err = c.mongoDAO.UpdateUsersStatus(ctx, []string{requestData.UserId}, models.UserStatusInGame)
	if err != nil {
		return err
	}

	c.gameServer.Mutex.Lock()
	c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId] = models.PlayerJoinedStatus

	// if channel is initialized to listen for any player joining the game
	if c.gameServer.Parties[requestData.PartyId].PlayerStatusUpdateMsg != nil {
		c.gameServer.Parties[requestData.PartyId].PlayerStatusUpdateMsg <- requestData.UserId + " has " + string(models.PlayerJoinedStatus) + " the party"
	}

	c.gameServer.Mutex.Unlock()

	return nil
}
