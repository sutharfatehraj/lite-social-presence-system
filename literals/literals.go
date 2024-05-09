package literals

const (
	EmptyString = ""

	// MongoDB
	Database            = "social-presence-system"
	UsersCollection     = "users"
	FriendsCollection   = "friends"
	GamePartyCollection = "gameparty"

	// MongoDB operators
	MongoOr       = "$or"
	MongoIn       = "$in"
	MongoSet      = "$set"
	MongoExpr     = "$expr"
	MongoLessThan = "$lt"
	MongoAdd      = "$add"
	MongoPush     = "$push"
	MongoPull     = "$pull"
	MongoEach     = "$each"
	MongoExists   = "$exists"

	// MongoDB fields
	MongoID          = "_id"
	MongoUserId      = "userId"
	MongoFriendId    = "friendId"
	MongoStatus      = "status"
	MongoRequestedBy = "requestedBy"
	MongoRequestedOn = "requestedOn"
	MongoCreatedBy   = "createdBy"
	MongoStartTime   = "startTime"
	MongoDuration    = "duration"

	MongoGamePartyInvitees = "invitees"
	MongoGamePartyAccepted = "accepted"
	MongoGamePartyRejected = "rejected"
	MongoGamePartyPlayers  = "players"
	MongoGamePartyExited   = "exited"
	MongoGamePartyRemoved  = "removed"

	MongoPlayers = "players."
)
