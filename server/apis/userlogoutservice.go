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

type UserLogOutService interface {
	LogOutUser(ctx context.Context, requestData *models.UserLogOutRequestData) error
}

var userLogOutServiceStruct UserLogOutService
var userLogOutServiceOnce sync.Once

type userLogOutService struct {
	mongoDAO mongodao.MongoDAO
}

func InitUserLogOutService(mongodao mongodao.MongoDAO) UserLogOutService {
	userLogOutServiceOnce.Do(func() {
		userLogOutServiceStruct = &userLogOutService{
			mongoDAO: mongodao,
		}
	})
	return userLogOutServiceStruct
}

func GetUserLogOutServiceStruct() UserLogOutService {
	if userLogOutServiceStruct == nil {
		panic("User Log Out Service not initialized")
	}
	return userLogOutServiceStruct
}

func UserLogOutHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()

	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.UserLogOutResponseData{
			Success: success,
			Errors:  errStrings,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	requestData := &models.UserLogOutRequestData{}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("failed to read message for user log Out request: %v\n", err)
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	}
	err = json.Unmarshal(data, requestData)
	if err != nil {
		fmt.Printf("failed to unmarshal message for user log Out request: %v\n", err)
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

	svc := GetUserLogOutServiceStruct()
	err = svc.LogOutUser(ctx, requestData)
	if err != nil {
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	} else {
		responseStatusCode = http.StatusOK
	}
}

func (u userLogOutService) LogOutUser(ctx context.Context, requestData *models.UserLogOutRequestData) error {

	result, err := u.mongoDAO.UpdateUsersStatus(ctx, []string{requestData.UserId}, models.UserStatusOffline)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("invalid userId")
	}
	if result.ModifiedCount == 0 {
		return errors.New("failed to logout")
	}

	return nil
}
