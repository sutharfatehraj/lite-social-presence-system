package models

import (
	"time"
)

type FriendRequestStatus string

const (
	FriendshipStatusUndefined FriendRequestStatus = "undefined"
	FriendshipStatusPending   FriendRequestStatus = "pending"
	FriendshipStatusAccepted  FriendRequestStatus = "accepted"
	FriendshipStatusRejected  FriendRequestStatus = "rejected"
)

type GetFriendsResponse struct {
	Success bool     `json:"success"`
	Friends []*User  `json:"friends,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type Friends struct {
	Id          string              `bson:"_id" json:"id"`
	UserId      string              `bson:"userId" json:"userId"`
	FriendId    string              `bson:"friendId" json:"friendId"`
	Status      FriendRequestStatus `bson:"status" json:"status"`
	RequestedBy string              `bson:"requestedBy" json:"requestedBy"` // userId of the user who requested the friendship
	RequestedOn time.Time           `bson:"requestedOn" json:"requestedOn"`
	// UpdatedOn   time.Time           `bson:"updatedOn" json:"updatedOn"` // TODO: add this in future
}

type SendFriendRequestData struct {
	UserId    string   `json:"userId"`
	FriendIds []string `json:"friendIds"`
}

type SendFriendRequestResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type HandleFriendRequestData struct {
	UserId    string              `json:"userId"`
	FriendIds []string            `json:"friendIds"`
	Status    FriendRequestStatus `json:"status"`
}

type HandleFriendRequestResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}

type RemoveFriendsRequestData struct {
	UserId    string   `json:"userId"`
	FriendIds []string `json:"friendIds"`
}

type RemoveFriendRequestResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}
