package models

import (
	"time"
)

type UserCredentials struct {
	ID       string `bson:"_id" json:"userid"` // userId is the primary key
	Password string `bson:"password" json:"-"` // struct tag '-' removes that field from getting printed anywhere
}

// user collection fields
type User struct {
	ID    string `bson:"_id" json:"userid"` // userId is the primary key
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Level string `bson:"level" json:"level"` // this field can be used on UI side to show some kind of symbol with the player
}

type GetFriendsResponse struct {
	Success bool     `json:"success"`
	Friends []*User  `json:"friends,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type FriendRequestStatus string

const (
	StatusPending  FriendRequestStatus = "pending"
	StatusAccepted FriendRequestStatus = "accepted"
	StatusRejected FriendRequestStatus = "rejected"
)

// accepted/rejected/pending

type Friends struct {
	Id          string              `bson:"_id" json:"id"`
	UserId      string              `bson:"userId" json:"userId"`
	FriendId    string              `bson:"friendId" json:"friendId"`
	Status      FriendRequestStatus `bson:"status" json:"status"`
	RequestedBy string              `bson:"requestedBy" json:"requestedBy"` // userId of the user who requested the friendship
	RequestedOn time.Time           `bson:"requestedOn" json:"requestedOn"`
	// UpdatedOn   time.Time           `bson:"updatedOn" json:"updatedOn"` // add this?
}

type SendFriendRequest struct {
	UserId    string   `json:"userid"`
	FriendIds []string `json:"friendIds"`
}

type SendFriendRequestResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type HandleFriendRequest struct {
	UserId    string              `json:"userid"`
	FriendIds []string            `json:"friendIds"`
	Status    FriendRequestStatus `json:"status"`
}

type HandleFriendRequestResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type RemoveFriendsRequestData struct {
	UserId    string   `json:"userid"`
	FriendIds []string `json:"friendIds"`
}

type RemoveFriendRequestResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}
