package common

import (
	"context"
	"fmt"
	"lite-social-presence-system/models"
	"lite-social-presence-system/mongodao"
	"time"
)

var PartyDuration time.Duration = 900000 * time.Millisecond // 15 minutes = 900 seconds = 9,00,000 in milliseconds

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
