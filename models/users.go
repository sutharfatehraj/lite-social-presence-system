package models

type UserStatus string

const (
	UserStatusUndefined UserStatus = "undefined"
	UserStatusOffline   UserStatus = "offline"
	UserStatusIdle      UserStatus = "idle" // online but not in any game party
	UserStatusInGame    UserStatus = "in-game"
	// UserStatusSuspended UserStatus = "suspended"
)

// to keep track of users who are listening for player online update
type UserServer struct {
	UserDetails map[string]*UserDetails
}

type UserDetails struct {
	// FriendId              string      `json:"friendId"`
	FriendOnlineUpdateMsg chan string `json:"playerStatusUpdateMsg"`
}

type UserCredentials struct {
	ID       string `bson:"_id" json:"userId"` // userId is the primary key
	Password string `bson:"password" json:"-"` // struct tag '-' removes that field from getting printed anywhere
}

// user collection fields
type User struct {
	ID     string     `bson:"_id" json:"userId"` // userId is the primary key
	Name   string     `bson:"name" json:"name"`
	Email  string     `bson:"email" json:"email"`
	Level  string     `bson:"level" json:"level"`             // this field can be used on UI side to show some kind of symbol with the player
	Status UserStatus `bson:"status" json:"status,omitempty"` // user status
}

type UserLogInRequestData struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

type UserLogInResponseData struct {
	Success     bool     `json:"success"`
	Errors      []string `json:"errors,omitempty"`
	UserDetails *User    `json:"userDetails,omitempty"`
}

type UserLogOutRequestData struct {
	UserId string `json:"userId"`
}

type UserLogOutResponseData struct {
	Success bool     `json:"success"`
	Errors  []string `json:"errors,omitempty"`
}
