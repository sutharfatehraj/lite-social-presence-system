package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"net/http"
	"sync"
)

type GetUsersService interface {
	GetUserDetails(userId string) ([]*models.User, error)
}

var getUsersServiceStruct GetUsersService
var getUsersServiceOnce sync.Once

type getUsersService struct {
	mongoDAO mongodao.MongoDAO
}

func InitGetUsersService(mongodao mongodao.MongoDAO) GetUsersService {
	getUsersServiceOnce.Do(func() {
		getUsersServiceStruct = &getUsersService{
			mongoDAO: mongodao,
		}
	})
	return getUsersServiceStruct
}

func GetUsersServiceStruct() GetUsersService {
	if getUsersServiceStruct == nil {
		panic("getUsers Service not initialized")
	}
	return getUsersServiceStruct
}

// GET Friends
func GetUsers(w http.ResponseWriter, r *http.Request) {

	// Retrieve the id from the query parameter
	query := r.URL.Query()
	userId := query.Get("id")
	fmt.Println("Request data: ", userId)

	var responseStatusCode int
	var result models.GetFriendsResponse

	defer func() {
		/*
			json.Marshal:
				useful if you need the JSON for other purposes before writing it to the response,
				such as logging or manipulating the data.
				Requires an extra step to write the data to http.ResponseWriter using w.Write(response)
				Less efficient as it requires an extra step of marshalling to a buffer before writing

		*/
		// response, err := json.Marshal(result)

		// if err != nil {
		// 	log.Printf("Unable to marshal the response : %v", err)
		// 	responseStatusCode = http.StatusInternalServerError
		// }
		// w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)

		// w.Write(response) // should always be at last

		/*
			Alternative, json.NewEncoder:
				directly writes to http.ResponseWriter.
				More efficient as it does not require a separate buffer
		*/
		json.NewEncoder(w).Encode(result) // should always be at last

		// In both the approaches, difference in performance is negligible
	}()

	if userId == literals.EmptyString {
		fmt.Println("no user ID passed")
		err := errors.New("no user ID passed")

		responseStatusCode = http.StatusBadRequest
		result = models.GetFriendsResponse{
			Success: false,
			Errors:  []string{err.Error()},
		}
	} else {
		success := true
		var errStrings []string

		svc := GetUsersServiceStruct()
		users, err := svc.GetUserDetails(userId)

		if err != nil {
			success = false
			responseStatusCode = http.StatusInternalServerError
			errStrings = append(errStrings, err.Error())
		} else {
			fmt.Printf("Retrieved: %+v", users)
			responseStatusCode = http.StatusOK
		}

		result = models.GetFriendsResponse{
			Success: success,
			Friends: users,
			Errors:  errStrings,
		}
	}
}

func (u getUsersService) GetUserDetails(userId string) ([]*models.User, error) {

	ctx := context.TODO()

	var users []*models.User
	var err error

	users, err = u.mongoDAO.GetUserDetails(ctx, []string{userId})
	// handle the errors.
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return users, nil

}

/* Dummy data
[]*models.User{{
	ID:    "123",
	Name:  "FJ",
	Email: "fj@xmail.com",
	Level: "amateur",
}}
*/
