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

type ExitGamePartyService interface {
	ValidateRequest(ctx context.Context, requestData *models.ExitGamePartyRequestData) []string
	ExitGameParty(ctx context.Context, requestData *models.ExitGamePartyRequestData) error
}

var exitGamePartyServiceStruct ExitGamePartyService
var exitGamePartyServiceOnce sync.Once

type exitGamePartyService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitExitGamePartyService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) ExitGamePartyService {
	exitGamePartyServiceOnce.Do(func() {
		exitGamePartyServiceStruct = &exitGamePartyService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return exitGamePartyServiceStruct
}

func GetExitGamePartyService() ExitGamePartyService {
	if exitGamePartyServiceStruct == nil {
		panic("ExitGameParty Service not initialized")
	}
	return exitGamePartyServiceStruct
}

func (c exitGamePartyService) ValidateRequest(ctx context.Context, requestData *models.ExitGamePartyRequestData) []string {
	var errs []error
	var errorString []string

	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	}

	if requestData.PartyId == literals.EmptyString {
		errs = append(errs, errors.New("empty partyId in the request data"))
	}

	if errs == nil {
		// user should be present in party as a player and should have status as joined
		if _, ok := c.gameServer.Parties[requestData.PartyId]; !ok {
			errs = append(errs, errors.New("invalid partyId in the request data"))
		} else if c.gameServer.Parties[requestData.PartyId].Players != nil {
			if playerStatus, ok := c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId]; ok {
				// player can exit if his current status is "joined"
				if playerStatus != models.PlayerJoinedStatus {
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

func ExitGamePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.ExitGamePartyResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.ExitGamePartyRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read exit game party message: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal exit game party message : %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetExitGamePartyService()

	errStrings = svc.ValidateRequest(ctx, requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.ExitGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to join the game party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c exitGamePartyService) ExitGameParty(ctx context.Context, requestData *models.ExitGamePartyRequestData) error {

	err := c.mongoDAO.UpdatePlayersDecisionForGameParty(ctx, requestData.PartyId, []string{requestData.UserId}, models.PlayerExitedStatus)
	if err != nil {
		return err
	}

	// update user status to idle
	_, err = c.mongoDAO.UpdateUsersStatus(ctx, []string{requestData.UserId}, models.UserStatusIdle)
	if err != nil {
		return err
	}

	c.gameServer.Mutex.Lock()
	c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId] = models.PlayerExitedStatus
	c.gameServer.Mutex.Unlock()

	return nil
}
