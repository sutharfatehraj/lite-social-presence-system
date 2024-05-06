package literals

const (
	EmptyString       = ""
	Database          = "social-presence-system"
	UsersCollection   = "users"
	FriendsCollection = "friends"

	// Mongo operators
	MongoOr  = "$or"
	MongoIn  = "$in"
	MongoSet = "$set"

	// Mongo fields
	MongoID          = "_id"
	MongoUserId      = "userId"
	MongoFriendId    = "friendId"
	MongoStatus      = "status"
	MongoRequestedBy = "requestedBy"
	MongoRequestedOn = "requestedOn"
)
