package common

import (
	"context"
	"errors"
	"fmt"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"time"
)

// create gameServer file and push all that over there

// var PartyDuration time.Duration = 900000 * time.Millisecond // 15 minutes = 900 seconds = 9,00,000 in milliseconds
var PartyDuration time.Duration = 900000000 * time.Millisecond // 10 days

func NewGameServer(mgDAO mongodao.MongoDAO) (*models.GameServer, error) {
	var gameParties []*models.GameParty
	var err error

	// fetch active game parties
	gameParties, err = mgDAO.FetchActiveGameParties(context.TODO())
	if err != nil {
		fmt.Println("Error finding active game parties", err)
		return nil, err
	}
	if gameParties == nil {
		return &models.GameServer{
			Parties: make(map[string]*models.GameParty),
		}, nil
	}

	parties := make(map[string]*models.GameParty)

	for _, gameParty := range gameParties {
		parties[gameParty.PartyId] = gameParty
	}

	return &models.GameServer{
		Parties: parties,
	}, nil
}

func NewUserServer() *models.UserServer {
	return &models.UserServer{
		UserDetails: make(map[string]*models.UserDetails),
	}

}

// remove data from one slice and append to the other
func RemoveAndAppendSlice(dataToRemove string, s1 []string, s2 []string) ([]string, []string, error) {

	indexOfDataToRemove := -1
	for index, data := range s1 {
		if data == dataToRemove {
			indexOfDataToRemove = index
			break
		}
	}
	if indexOfDataToRemove == -1 {
		return nil, nil, errors.New("data not found in the slice")
	}

	s1 = append(s1[:indexOfDataToRemove], s1[indexOfDataToRemove+1:]...)
	s2 = append(s2, dataToRemove)

	return s1, s2, nil
}
