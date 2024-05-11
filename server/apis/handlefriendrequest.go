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

type HandleFriendRequestService interface {
	ValidateRequest(requestData *models.HandleFriendRequestData) []string
	UpdateFriendRequestStatus(ctx context.Context, requestData *models.HandleFriendRequestData) error
}

var handleFriendRequestStruct HandleFriendRequestService
var handleFriendRequestOnce sync.Once

type handleFriendRequest struct {
	mongoDAO mongodao.MongoDAO
}

func InitHandleFriendRequestService(mongodao mongodao.MongoDAO) HandleFriendRequestService {
	handleFriendRequestOnce.Do(func() {
		handleFriendRequestStruct = &handleFriendRequest{
			mongoDAO: mongodao,
		}
	})
	return handleFriendRequestStruct
}

func GetHandleFriendRequestService() HandleFriendRequestService {
	if handleFriendRequestStruct == nil {
		panic("HandleFriendRequestService Service not initialized")
	}
	return handleFriendRequestStruct
}

func (h handleFriendRequest) ValidateRequest(requestData *models.HandleFriendRequestData) []string {

	var errs []error
	var errorString []string

	//  user Id should not be empty
	if requestData.UserId == literals.EmptyString {
		errs = append(errs, errors.New("empty userId in the request data"))
	}

	// status should be accepted/rejected
	if requestData.Status != models.FriendshipStatusAccepted && requestData.Status != models.FriendshipStatusRejected {
		errs = append(errs, errors.New("status not qual to accepted/rejected in the request data"))
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

// handle-friend-request
func HandleFriendRequest(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.HandleFriendRequestResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.HandleFriendRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read message for handle-freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal message for handle-freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	svc := GetHandleFriendRequestService()

	errStrings = svc.ValidateRequest(requestData)
	if errStrings != nil {
		success = false
		responseStatusCode = http.StatusBadRequest
		return
	}

	err = svc.UpdateFriendRequestStatus(ctx, requestData)
	if err != nil {
		fmt.Printf("failed to store handle-freindship request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

}

func (h handleFriendRequest) UpdateFriendRequestStatus(ctx context.Context, requestData *models.HandleFriendRequestData) error {

	var allUserIds []string
	allUserIds = append(allUserIds, requestData.UserId)
	allUserIds = append(allUserIds, requestData.FriendIds...)

	// if all are present in DB, continue else return error
	_, err := h.mongoDAO.GetUserDetails(ctx, allUserIds)
	if err != nil {
		return err
	}

	/*
		Accepting/rejecting can be done by only the userId to whom the friend request was sent
		This will not be an issue, if the API is called with correct data from the client side.
				Ex, 114 sent friend request to 113

				now, only 113 should be allowed to accept/reject the friend-request

				from db, fetch all docs from friends collection where
				userId = requestedUserId,
				friendId = requestedFriendId,
				requestedFriendId = db requestedBy

				If count of results > 0, user can update the friend request status

	*/

	err = h.mongoDAO.UpdateFriendRequestsStatus(ctx, requestData.UserId, requestData.FriendIds, requestData.Status)
	if err != nil {
		return err
	}

	return nil
}
