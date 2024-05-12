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

type GetFriendsService interface {
	GetFriends(ctx context.Context, userId string) ([]*models.User, error)
}

var getFriendsServiceStruct GetFriendsService
var getFriendsServiceOnce sync.Once

type getFriendsService struct {
	mongoDAO mongodao.MongoDAO
}

func InitGetFriendsService(mongodao mongodao.MongoDAO) GetFriendsService {
	getFriendsServiceOnce.Do(func() {
		getFriendsServiceStruct = &getFriendsService{
			mongoDAO: mongodao,
		}
	})
	return getFriendsServiceStruct
}

func GetFriendsServiceStruct() GetFriendsService {
	if getFriendsServiceStruct == nil {
		panic("getFriends Service not initialized")
	}
	return getFriendsServiceStruct
}

// GET Friends
func GetFriendsHandler(w http.ResponseWriter, r *http.Request) {

	ctx := context.TODO()

	var friends []*models.User
	success := true
	var responseStatusCode int = http.StatusOK
	var errStrings []string
	var err error

	defer func() {
		result := models.GetFriendsResponse{
			Success: success,
			Errors:  errStrings,
			Friends: friends,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseStatusCode)
		json.NewEncoder(w).Encode(result)
	}()

	// Retrieve id from the query parameter
	query := r.URL.Query()
	userId := query.Get("id")
	fmt.Println("Request data: ", userId)

	if userId == literals.EmptyString {
		fmt.Println("no user ID passed")
		err := errors.New("no user ID passed")

		success = false
		responseStatusCode = http.StatusBadRequest
		errStrings = append(errStrings, err.Error())
		return
	}

	svc := GetFriendsServiceStruct()
	friends, err = svc.GetFriends(ctx, userId)
	if err != nil {
		success = false
		responseStatusCode = http.StatusInternalServerError
		errStrings = append(errStrings, err.Error())
		return
	} else {
		fmt.Printf("Found %d friends\n", len(friends))
		responseStatusCode = http.StatusOK
	}

}

func (f getFriendsService) GetFriends(ctx context.Context, userId string) ([]*models.User, error) {

	var friends []*models.User
	var err error

	friends, err = f.mongoDAO.GetFriendsDetails(ctx, userId)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return friends, nil
}
