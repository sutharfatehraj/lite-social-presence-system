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

type RemoveUsersFromGamePartyService interface {
	ValidateRequest(ctx context.Context, requestData *models.RemoveUsersFromGamePartyRequestData) []string
	RemoveUsersFromGameParty(ctx context.Context, requestData *models.RemoveUsersFromGamePartyRequestData) error
}

var removeUsersFromGamePartyServiceStruct RemoveUsersFromGamePartyService
var removeUsersFromGamePartyServiceOnce sync.Once

type removeUsersFromGamePartyService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitRemoveUsersFromGamePartyService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) RemoveUsersFromGamePartyService {
	removeUsersFromGamePartyServiceOnce.Do(func() {
		removeUsersFromGamePartyServiceStruct = &removeUsersFromGamePartyService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return removeUsersFromGamePartyServiceStruct
}

func GetRemoveUsersFromGamePartyService() RemoveUsersFromGamePartyService {
	if removeUsersFromGamePartyServiceStruct == nil {
		panic("RemoveUsersFromGameParty Service not initialized")
	}
	return removeUsersFromGamePartyServiceStruct
}

func (c removeUsersFromGamePartyService) ValidateRequest(ctx context.Context, requestData *models.RemoveUsersFromGamePartyRequestData) []string {
	var errs []error
	var errorString []string

	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	} else if c.gameServer.Parties[requestData.PartyId] != nil && requestData.UserId != c.gameServer.Parties[requestData.PartyId].CreatedBy {
		errs = append(errs, errors.New("party "+requestData.PartyId+" not created by "+requestData.UserId))
	}

	// friend Ids should not be empty
	if len(requestData.FriendIds) == 0 {
		errs = append(errs, errors.New("no friendIds found in the request data"))
	} else {
		friendIdCount := make(map[string]int)
		emptyFriendId := false
		for _, friendId := range requestData.FriendIds {
			if friendId == literals.EmptyString {
				emptyFriendId = true
				break
			}
			friendIdCount[friendId]++
		}
		if emptyFriendId {
			errs = append(errs, errors.New("found empty friendId in the request data"))
		} else {
			// friend Ids should not be repeated more than once
			for friendId, count := range friendIdCount {
				if count > 1 {
					errs = append(errs, errors.New("friendId "+friendId+" sent "+fmt.Sprint(count)+" times in the request data"))
				}
			}
		}
	}

	if requestData.PartyId == literals.EmptyString {
		errs = append(errs, errors.New("empty partyId in the request data"))
	}

	// check party data only if userId and friendIds are correct
	if errs == nil {
		if requestData.PartyId == literals.EmptyString {
			errs = append(errs, errors.New("empty partyId in the request data"))
		} else {
			if _, ok := c.gameServer.Parties[requestData.PartyId]; !ok {
				errs = append(errs, errors.New("invalid partyId in the request data"))
			} else if c.gameServer.Parties[requestData.PartyId].Players != nil {
				// players present in the game party
				for _, playerId := range requestData.FriendIds {
					if playerStatus, ok := c.gameServer.Parties[requestData.PartyId].Players[playerId]; ok {
						// playerId already present
						// can be removed if player status is 'joined'
						if playerStatus != models.PlayerJoinedStatus {
							// cannot remove this player
							errs = append(errs, errors.New("player "+playerId+" cannot be removed. Has status: "+string(playerStatus)))
						}
					}
				}
			}
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

func RemoveFromGamePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.RemoveUsersFromGamePartyResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.RemoveUsersFromGamePartyRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read remove from game party message: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal remove from game party message : %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetRemoveUsersFromGamePartyService()

	errStrings = svc.ValidateRequest(ctx, requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.RemoveUsersFromGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to join the game party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c removeUsersFromGamePartyService) RemoveUsersFromGameParty(ctx context.Context, requestData *models.RemoveUsersFromGamePartyRequestData) error {

	err := c.mongoDAO.UpdatePlayersDecisionForGameParty(ctx, requestData.PartyId, requestData.FriendIds, models.PlayerRemovedStatus)
	if err != nil {
		return err
	}

	// update the users status to idle
	_, err = c.mongoDAO.UpdateUsersStatus(ctx, requestData.FriendIds, models.UserStatusIdle)
	if err != nil {
		return err
	}

	c.gameServer.Mutex.Lock()
	for _, playerId := range requestData.FriendIds {
		c.gameServer.Parties[requestData.PartyId].Players[playerId] = models.PlayerRemovedStatus
	}
	c.gameServer.Mutex.Unlock()

	return nil
}
