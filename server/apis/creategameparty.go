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
	"lite-social-presence-system/server/common"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CreateGamePartyService interface {
	CreateAndStoreGameParty(ctx context.Context, requestData *models.CreateGamePartyRequestData) (string, error)
}

var createGamePartyServiceStruct CreateGamePartyService
var createGamePartyServiceOnce sync.Once

type createGamePartyService struct {
	gameServer *models.GameServer
	mongoDAO   mongodao.MongoDAO
}

func InitCreateGamePartyService(gameSrvr *models.GameServer, mongodao mongodao.MongoDAO) CreateGamePartyService {
	createGamePartyServiceOnce.Do(func() {
		createGamePartyServiceStruct = &createGamePartyService{
			gameServer: gameSrvr,
			mongoDAO:   mongodao,
		}
	})
	return createGamePartyServiceStruct
}

func GetCreateGamePartyService() CreateGamePartyService {
	if createGamePartyServiceStruct == nil {
		panic("CreateGameParty Service not initialized")
	}
	return createGamePartyServiceStruct
}

func CreateGameParty(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error
	var partyId string

	defer func() {
		result := models.CreateGamePartyResponseData{
			Success: success,
			PartyId: partyId,
			Errors:  errStrings,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.CreateGamePartyRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read message for create party request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal message for create party request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	if requestData.UserId == literals.EmptyString {
		fmt.Println("no user ID passed")
		err := errors.New("no user ID passed")

		success = false
		responseStatusCode = http.StatusBadRequest
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetCreateGamePartyService()
	partyId, err = svc.CreateAndStoreGameParty(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to create party: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
}

func (c createGamePartyService) CreateAndStoreGameParty(ctx context.Context, requestData *models.CreateGamePartyRequestData) (string, error) {

	partyId := uuid.NewString()
	gameParty := &models.GameParty{
		PartyId:   partyId,
		CreatedBy: requestData.UserId,
		StartTime: time.Now(),
		Duration:  common.PartyDuration,
		Status:    models.GamePartyStatusActive,
	}

	// store game party in DB
	err := c.mongoDAO.CreateGameParty(ctx, gameParty)
	if err != nil {
		return literals.EmptyString, err
	}

	// update the user status to "in-game"
	_, err = c.mongoDAO.UpdateUsersStatus(ctx, []string{requestData.UserId}, models.UserStatusInGame)
	if err != nil {
		return literals.EmptyString, err
	}

	c.gameServer.Mutex.Lock()
	c.gameServer.Parties[partyId] = gameParty
	c.gameServer.Mutex.Unlock()
	return partyId, nil
}
