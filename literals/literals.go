package literals

const (
	RestAPIServerAddress = "127.0.0.1:8081"

	// gRPC
	GRPCNetwork       = "tcp"
	GRPCServerAddress = "127.0.0.1:8083"

	EmptyString = ""

	// MongoDB
	Database            = "social-presence-system"
	UsersCollection     = "users"
	FriendsCollection   = "friends"
	GamePartyCollection = "gameparty"
	UserCredsCollection = "usercreds"

	// MongoDB operators
	MongoOr       = "$or"
	MongoIn       = "$in"
	MongoAnd      = "$and"
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
	MongoPassword    = "password"
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

	MongoPlayersDotAccess = "players."
)
