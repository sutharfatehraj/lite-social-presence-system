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

type HandleGamePartyInviteService interface {
	ValidateRequest(ctx context.Context, requestData *models.HandleGamePartyInviteRequestData) []string
	HandleInvitationToGameParty(ctx context.Context, requestData *models.HandleGamePartyInviteRequestData) error
}

var handleGamePartyInviteServiceStruct HandleGamePartyInviteService
var handleGamePartyInviteServiceOnce sync.Once

type handleGamePartyInviteService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitHandleGamePartyInviteService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) HandleGamePartyInviteService {
	handleGamePartyInviteServiceOnce.Do(func() {
		handleGamePartyInviteServiceStruct = &handleGamePartyInviteService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return handleGamePartyInviteServiceStruct
}

func GetHandleGamePartyInviteService() HandleGamePartyInviteService {
	if handleGamePartyInviteServiceStruct == nil {
		panic("HandleGamePartyInvite Service not initialized")
	}
	return handleGamePartyInviteServiceStruct
}

func (c handleGamePartyInviteService) ValidateRequest(ctx context.Context, requestData *models.HandleGamePartyInviteRequestData) []string {
	var errs []error
	var errorString []string

	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	}

	if requestData.Status != models.PlayerAcceptedStatus && requestData.Status != models.PlayerRejectedStatus {
		errs = append(errs, errors.New("invalid status "+string(requestData.Status)+" in the request data"))
	}

	if requestData.PartyId == literals.EmptyString {
		errs = append(errs, errors.New("empty partyId in the request data"))
	}

	if errs == nil {
		// user should be present in party as a player and should have status as invited
		if _, ok := c.gameServer.Parties[requestData.PartyId]; !ok {
			errs = append(errs, errors.New("invalid partyId in the request data"))
		} else if c.gameServer.Parties[requestData.PartyId].Players != nil {
			if playerStatus, ok := c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId]; ok {
				// player can accept or reject if his current status is "invited"
				if playerStatus != models.PlayerInvitedStatus {
					errs = append(errs, errors.New("player "+requestData.UserId+" has current status: "+string(playerStatus)+". cannot update decision to "+string(requestData.Status)))
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

func HandleGamePartyInviteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.HandleGamePartyInviteResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.HandleGamePartyInviteRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read handle game party message: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal handle game party message : %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetHandleGamePartyInviteService()

	errStrings = svc.ValidateRequest(ctx, requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.HandleInvitationToGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to handle invitation to the game party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c handleGamePartyInviteService) HandleInvitationToGameParty(ctx context.Context, requestData *models.HandleGamePartyInviteRequestData) error {

	err := c.mongoDAO.UpdatePlayersDecisionForGameParty(ctx, requestData.PartyId, []string{requestData.UserId}, requestData.Status)
	if err != nil {
		return err
	}

	c.gameServer.Mutex.Lock()
	c.gameServer.Parties[requestData.PartyId].Players[requestData.UserId] = requestData.Status
	c.gameServer.Mutex.Unlock()

	return nil
}
