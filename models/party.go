package models

import (
	"sync"
	"time"
)

type GamePartyStatus string

const (
	GamePartyStatusOver   GamePartyStatus = "over"
	GamePartyStatusActive GamePartyStatus = "active"
)

type GameParty struct {
	PartyId   string          `bson:"_id" json:"partyId"`         // unique game party identifier
	CreatedBy string          `bson:"createdBy" json:"createdBy"` // user that created the party
	StartTime time.Time       `bson:"startTime" json:"startTime"` // time when party was created
	Duration  time.Duration   `bson:"duration" json:"duration"`   // duration for which the party is created
	Invitees  []string        `bson:"invitees" json:"invitees"`   // users who are invited to the party. They can later accept/reject the invitation
	Accepted  []string        `bson:"accepted" json:"accepted"`   // users who have accepted the invitation. They can later join the party
	Rejected  []string        `bson:"rejected" json:"rejected"`   // users who have rejected the invitation.
	Players   []string        `bson:"players" json:"players"`     // users who have joined the party
	Exited    []string        `bson:"exited" json:"exited"`       // players who have left the party, can rejoin again if the party is still ON?
	Removed   []string        `bson:"removed" json:"removed"`     // players who have been removed from the group
	Status    GamePartyStatus `bson:"status" json:"status"`       // status of the game party
}

type GameServer struct {
	Parties map[string]*GameParty
	Mutex   sync.Mutex
}

type CreateGamePartyRequestData struct {
	UserId string `json:"userid"` // unique identifier of the user creating the party
}

type CreateGamePartyResponseData struct {
	Success bool     `json:"success"`
	PartyId string   `json:"partyId,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}
