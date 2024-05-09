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

type InviteToGamePartyService interface {
	ValidateRequest(ctx context.Context, requestData *models.InviteToGamePartyRequestData) []string
	StoreInvitationToGameParty(ctx context.Context, requestData *models.InviteToGamePartyRequestData) error
}

var inviteToGamePartyServiceStruct InviteToGamePartyService
var inviteToGamePartyServiceOnce sync.Once

type inviteToGamePartyService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitInviteToGamePartyService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) InviteToGamePartyService {
	inviteToGamePartyServiceOnce.Do(func() {
		inviteToGamePartyServiceStruct = &inviteToGamePartyService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return inviteToGamePartyServiceStruct
}

func GetInviteToGamePartyService() InviteToGamePartyService {
	if inviteToGamePartyServiceStruct == nil {
		panic("InviteToGameParty Service not initialized")
	}
	return inviteToGamePartyServiceStruct
}

func (c inviteToGamePartyService) ValidateRequest(ctx context.Context, requestData *models.InviteToGamePartyRequestData) []string {
	var errs []error
	var errorString []string

	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	} else if c.gameServer.Parties[requestData.PartyId] != nil && c.gameServer.Parties[requestData.PartyId].CreatedBy != literals.EmptyString && requestData.UserId != c.gameServer.Parties[requestData.PartyId].CreatedBy {
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

	// check party data only if userId and friendIds are correct
	if errs == nil {
		if requestData.PartyId == literals.EmptyString {
			errs = append(errs, errors.New("empty partyId in the request data"))
		} else {
			if _, ok := c.gameServer.Parties[requestData.PartyId]; !ok {
				errs = append(errs, errors.New("invalid partyId in the request data"))
			} else if c.gameServer.Parties[requestData.PartyId].Players != nil {
				// players present
				for _, playerId := range requestData.FriendIds {
					if playerStatus, ok := c.gameServer.Parties[requestData.PartyId].Players[playerId]; ok {
						// playerId already present
						// can be invited again if player status is rejected/exited/removed
						if playerStatus != models.PlayerExitedStatus && playerStatus != models.PlayerRejectedStatus && playerStatus != models.PlayerRemovedStatus {
							// cannot invite this player
							errs = append(errs, errors.New("player "+playerId+" cannot be invited. Has status: "+string(playerStatus)))
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

func InviteToGamePartyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.InviteToGamePartyResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.InviteToGamePartyRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read invite to game party message: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal invite to game party message : %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetInviteToGamePartyService()

	errStrings = svc.ValidateRequest(ctx, requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.StoreInvitationToGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to store invitation to the game party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c inviteToGamePartyService) StoreInvitationToGameParty(ctx context.Context, requestData *models.InviteToGamePartyRequestData) error {

	areFriends, err := c.mongoDAO.CheckFriendship(ctx, requestData.UserId, requestData.FriendIds)
	if err != nil {
		return err
	}
	if areFriends {

		// append the the userId to the invitees array
		err = c.mongoDAO.AddInviteesToGamePartyCollection(ctx, requestData.PartyId, requestData.FriendIds)
		if err != nil {
			return err
		}

		c.gameServer.Mutex.Lock()
		if c.gameServer.Parties[requestData.PartyId].Players == nil {
			c.gameServer.Parties[requestData.PartyId].Players = make(map[string]models.GamePartyPlayerStatus)
		}
		for _, playerId := range requestData.FriendIds {
			c.gameServer.Parties[requestData.PartyId].Players[playerId] = models.PlayerInvitedStatus
		}
		c.gameServer.Mutex.Unlock()

	}

	return nil
}
