package models

import (
	"sync"
	"time"
)

type GamePartyStatus string

const (
	GamePartyStatusUndefined GamePartyStatus = "undefined"
	GamePartyStatusOver      GamePartyStatus = "over"
	GamePartyStatusActive    GamePartyStatus = "active"
)

type GamePartyPlayerStatus string

const (
	PlayerStatusUndefined GamePartyPlayerStatus = "undefined"
	PlayerInvitedStatus   GamePartyPlayerStatus = "invited"  // users who are invited to the party. They can later accept/reject the invitation
	PlayerAcceptedStatus  GamePartyPlayerStatus = "accepted" // users who have accepted the invitation. They can later join the party
	PlayerRejectedStatus  GamePartyPlayerStatus = "rejected" // users who have rejected the invitation. Can be invited again
	PlayerJoinedStatus    GamePartyPlayerStatus = "joined"   // users who have joined the party
	PlayerExitedStatus    GamePartyPlayerStatus = "exited"   // players who have left the party. Can be invited again
	PlayerRemovedStatus   GamePartyPlayerStatus = "removed"  // players who have been removed from the game party. Can be invited again
)

type GameParty struct {
	PartyId   string          `bson:"_id" json:"partyId"`         // unique game party identifier
	CreatedBy string          `bson:"createdBy" json:"createdBy"` // user that created the party
	StartTime time.Time       `bson:"startTime" json:"startTime"` // time when party was created
	Duration  time.Duration   `bson:"duration" json:"duration"`   // duration for which the party is created
	Status    GamePartyStatus `bson:"status" json:"status"`       // status of the game party
	Players   map[string]GamePartyPlayerStatus
}

type GameServer struct {
	Parties map[string]*GameParty
	Mutex   sync.Mutex
}

type CreateGamePartyRequestData struct {
	UserId string `json:"userId"` // unique identifier of the user creating the party
}

type CreateGamePartyResponseData struct {
	Success bool     `json:"success"`
	PartyId string   `json:"partyId,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type InviteToGamePartyRequestData struct {
	PartyId   string   `json:"partyId"`
	UserId    string   `json:"userId"`
	FriendIds []string `json:"friendIds"`
}

type InviteToGamePartyResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type HandleGamePartyInviteRequestData struct {
	PartyId string                `json:"partyId"`
	UserId  string                `json:"userId"`
	Status  GamePartyPlayerStatus `json:"status"` // to accept/reject the game party invitation
}

type HandleGamePartyInviteResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type JoinGamePartyRequestData struct {
	PartyId string `json:"partyId"`
	UserId  string `json:"userId"`
}

type JoinGamePartyResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type ExitGamePartyRequestData struct {
	PartyId string `json:"partyId"`
	UserId  string `json:"userId"`
}

type ExitGamePartyResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type RemoveUserFromGamePartyRequestData struct {
	PartyId   string   `json:"partyId"`
	UserId    string   `json:"userId"`    // user Id of the user who created the party
	PlayerIds []string `json:"playerIds"` // must have 'joined' status for him to be removed
}

type RemoveUserFromGamePartyResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}
