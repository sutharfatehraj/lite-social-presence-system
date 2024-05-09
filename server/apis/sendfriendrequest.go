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

type SendFriendRequestService interface {
	ValidateRequest(requestData *models.SendFriendRequestData) []string
	AddFriendRequest(ctx context.Context, requestData *models.SendFriendRequestData) error
}

var sendFriendRequestStruct SendFriendRequestService
var sendFriendRequestOnce sync.Once

type sendFriendRequest struct {
	mongoDAO mongodao.MongoDAO
}

func InitSendFriendRequestService(mongodao mongodao.MongoDAO) SendFriendRequestService {
	sendFriendRequestOnce.Do(func() {
		sendFriendRequestStruct = &sendFriendRequest{
			mongoDAO: mongodao,
		}
	})
	return sendFriendRequestStruct
}

func SendFriendRequestServiceStruct() SendFriendRequestService {
	if sendFriendRequestStruct == nil {
		panic("SendFriendRequestService Service not initialized")
	}
	return sendFriendRequestStruct
}
func (r sendFriendRequest) ValidateRequest(requestData *models.SendFriendRequestData) []string {

	var errs []error
	var errorString []string

	//  user Id should not be empty
	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
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

	if len(errs) > 0 {
		for _, err := range errs {
			errorString = append(errorString, err.Error())
		}
		return errorString
	}

	return nil
}

// Send Friend Request
func SendFriendRequestHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.SendFriendRequestResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.SendFriendRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read message for freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal message for freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := SendFriendRequestServiceStruct()

	errStrings = svc.ValidateRequest(requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}
	err = svc.AddFriendRequest(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to store freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

}

func (s sendFriendRequest) AddFriendRequest(ctx context.Context, requestData *models.SendFriendRequestData) error {

	var allUserIds []string
	allUserIds = append(allUserIds, requestData.UserId)
	allUserIds = append(allUserIds, requestData.FriendIds...)

	// if all are present in DB, continue else return error
	_, err := s.mongoDAO.GetUserDetails(ctx, allUserIds)
	if err != nil {
		return err
	}

	err = s.mongoDAO.StoreFriendRequests(ctx, requestData.UserId, requestData.FriendIds)
	if err != nil {
		return err
	}

	return nil
}
