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

type UserLoginService interface {
	LogInUser(ctx context.Context, requestData *models.UserLogInRequestData) (*models.User, error)
}

var userLoginServiceStruct UserLoginService
var userLoginServiceOnce sync.Once

type userLoginService struct {
	mongoDAO   mongodao.MongoDAO
	userServer *models.UserServer
}

func InitUserLoginService(mongodao mongodao.MongoDAO, userSrvr *models.UserServer) UserLoginService {
	userLoginServiceOnce.Do(func() {
		userLoginServiceStruct = &userLoginService{
			mongoDAO:   mongodao,
			userServer: userSrvr,
		}
	})
	return userLoginServiceStruct
}

func GetUserLoginServiceStruct() UserLoginService {
	if userLoginServiceStruct == nil {
		panic("User Login Service not initialized")
	}
	return userLoginServiceStruct
}

func UserLogInHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()

	var userDetails *models.User
	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.UserLogInResponseData{
			Success:     success,
			Errors:      errStrings,
			UserDetails: userDetails,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.UserLogInRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read message for user login request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal message for user login request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}

	fmt.Printf("Request data: %+v\n", requestData)

	if requestData.UserId == literals.EmptyString {
		fmt.Println("no user ID passed")
		err := errors.New("no user ID passed")

		success = false
		responseStatusCode = http.StatusBadRequest
		errStrings = append(errStrings, err.Error())
		return
	}

	svc := GetUserLoginServiceStruct()
	userDetails, err = svc.LogInUser(ctx, requestData)
	if err != nil {
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	} else {
		fmt.Printf("Retrieved: %+v", userDetails)
		responseStatusCode = http.StatusOK
	}
}

func (u userLoginService) LogInUser(ctx context.Context, requestData *models.UserLogInRequestData) (*models.User, error) {

	var users []*models.User
	var err error
	var isUserPresent bool

	// check user creds
	isUserPresent, err = u.mongoDAO.CheckUserCreds(ctx, requestData.UserId, requestData.Password)
	if err != nil {
		return nil, err
	}

	if isUserPresent {
		// update the status to 'idle'
		_, err = u.mongoDAO.UpdateUsersStatus(ctx, []string{requestData.UserId}, models.UserStatusIdle)
		if err != nil {
			return nil, err
		}

		users, err = u.mongoDAO.GetUserDetails(ctx, []string{requestData.UserId})
		if err != nil {
			return nil, err
		}
		go AsyncMsgPublishToFriend(ctx, u, requestData.UserId)
		return users[0], nil
	}
	return nil, err
}

func AsyncMsgPublishToFriend(ctx context.Context, u userLoginService, userId string) {
	// find this user's friends
	friends, friendFetchErr := u.mongoDAO.GetUserFriends(ctx, userId)
	if friendFetchErr != nil {
		return
	}
	for _, friend := range friends {
		if u.userServer != nil && u.userServer.UserDetails != nil && u.userServer.UserDetails[friend.FriendId] != nil && u.userServer.UserDetails[friend.FriendId].FriendOnlineUpdateMsg != nil {
			u.userServer.UserDetails[friend.FriendId].FriendOnlineUpdateMsg <- fmt.Sprintf("%v is now online", userId)
		}
	}

}
