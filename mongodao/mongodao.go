package mongodao

import (
	"context"
	"errors"
	"fmt"
	"lite-social-presence-system/literals"
	"lite-social-presence-system/models"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDAO interface {
	GetFriends(ctx context.Context, userId string) ([]*models.User, error)
	GetUserDetails(ctx context.Context, userIds []string) ([]*models.User, error)
	UpdateUsersStatus(ctx context.Context, userIds []string, status models.UserStatus) (*mongo.UpdateResult, error)
	StoreFriendRequests(ctx context.Context, userId string, friendIds []string) error
	UpdateFriendRequestsStatus(ctx context.Context, userId string, friendIds []string, status models.FriendRequestStatus) error
	RemoveFriends(ctx context.Context, userId string, friendIds []string) error

	// game party
	FetchActiveGameParties(ctx context.Context) ([]*models.GameParty, error)
	UpdateGamePartyStatus(ctx context.Context, partyIds []string, status models.GamePartyStatus) error
	CreateGameParty(ctx context.Context, gamePary *models.GameParty) error
	CheckFriendship(ctx context.Context, userId string, friendIds []string) (bool, error)
	AddInviteesToGamePartyCollection(ctx context.Context, partyId string, newInvitees []string) error
	UpdatePlayersDecisionForGameParty(ctx context.Context, partyId string, userIds []string, playerStatus models.GamePartyPlayerStatus) error

	// UpdatePlayerAndUserStatusForGameParty(ctx context.Context, partyId string, userId string, playerStatus models.GamePartyPlayerStatus, userStatus models.UserStatus) error
	// obsolete
	// PullAndPushDataInGamePartyCollection(ctx context.Context, partyId string, userId string, removeFrom string, addTo string) error
}

var mongoDAOStruct MongoDAO
var mongodaoOnce sync.Once

type mongoDAO struct {
	client  *mongo.Client
	databse *mongo.Database
}

func InitMongoDao(clnt *mongo.Client, db *mongo.Database) MongoDAO {
	mongodaoOnce.Do(func() {
		mongoDAOStruct = &mongoDAO{
			client:  clnt,
			databse: db,
		}
	})
	return mongoDAOStruct
}

func GetMongoDao() MongoDAO {
	if mongoDAOStruct == nil {
		panic("mongoDAO not initialized")
	}
	return mongoDAOStruct
}

/*
mongo.Client will be used for further database operation.
context.Context will be used set deadlines for process.
context.CancelFunc will be used to cancel context and resource associtated with it.
*/
func Connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	// set deadline for the process
	ctx, cancel := context.WithTimeout(context.Background(),
		60*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

// To close MongoDB connection and cancel context.
func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {

	// to cancel context
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

// To ping the mongoDB
func Ping(client *mongo.Client, ctx context.Context) error {

	// deadline of the Ping method will be determined by ctx
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	fmt.Println("MongoDB connected successfully")
	return nil
}

// Get all users who have accepted the friend request
func (m mongoDAO) GetFriends(ctx context.Context, userId string) ([]*models.User, error) {

	filter := bson.M{
		literals.MongoUserId: userId,
		literals.MongoStatus: models.FriendshipStatusAccepted,
	}

	cur, err := m.databse.Collection(literals.FriendsCollection).Find(ctx, filter)
	if err != nil {
		fmt.Println("Error occurred while calling friends. ", err)
		return nil, err
	}

	var friends []*models.Friends
	for cur.Next(ctx) {
		var friend models.Friends
		decodeErr := cur.Decode(&friend)
		if decodeErr != nil {
			fmt.Println(decodeErr)
			return nil, decodeErr
		}
		friends = append(friends, &friend)
	}

	if len(friends) == 0 {
		fmt.Println("No friends found")
		return nil, nil
	}

	var friendIds []string

	for _, friend := range friends {
		friendIds = append(friendIds, friend.FriendId)
	}

	return m.GetUserDetails(ctx, friendIds)
}

// Get user details
func (m mongoDAO) GetUserDetails(ctx context.Context, userIds []string) ([]*models.User, error) {

	filter := bson.M{
		literals.MongoID: bson.M{literals.MongoIn: userIds},
	}

	cur, err := m.databse.Collection(literals.UsersCollection).Find(ctx, filter)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}

	var users []*models.User
	for cur.Next(ctx) {
		var user models.User
		decodeErr := cur.Decode(&user)
		if decodeErr != nil {
			fmt.Println(decodeErr)
			return nil, decodeErr
		}
		users = append(users, &user)
	}

	if len(users) == 0 {
		fmt.Println("No users found")
		return nil, errors.New("no users found")
	} else if len(users) != len(userIds) {
		logrus.WithFields(logrus.Fields{
			literals.LLRequestedUserIds: userIds,
			literals.LLUsersFound:       (fmt.Sprintf("%+v", users)),
		}).Error("some users not found")

		return nil, errors.New("some users not found")
	}
	return users, err
}

func (m mongoDAO) UpdateUsersStatus(ctx context.Context, userIds []string, status models.UserStatus) (*mongo.UpdateResult, error) {

	filter := bson.M{
		literals.MongoID: bson.M{literals.MongoIn: userIds},
	}

	update := bson.M{
		literals.MongoSet: bson.M{
			literals.MongoStatus: status,
		},
	}

	result, err := m.databse.Collection(literals.UsersCollection).UpdateMany(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to update users status in the users collection. Err: %v\nUpdateResult: %v\n", err, result)
		return nil, err
	}
	return result, nil
}

func (m mongoDAO) StoreFriendRequests(ctx context.Context, userId string, friendIds []string) error {

	var docs []interface{}

	time := time.Now()

	for _, friendId := range friendIds {
		docs = append(docs, bson.M{
			literals.MongoID:          uuid.NewString(),
			literals.MongoUserId:      userId,
			literals.MongoStatus:      models.FriendshipStatusPending,
			literals.MongoRequestedBy: userId,
			literals.MongoFriendId:    friendId,
			literals.MongoRequestedOn: time,
		})

		docs = append(docs, bson.M{
			literals.MongoID:          uuid.NewString(),
			literals.MongoUserId:      friendId,
			literals.MongoStatus:      models.FriendshipStatusPending,
			literals.MongoRequestedBy: userId,
			literals.MongoFriendId:    userId,
			literals.MongoRequestedOn: time,
		})
	}

	result, err := m.databse.Collection(literals.FriendsCollection).InsertMany(ctx, docs)
	if err != nil {
		fmt.Printf("failed to insert friend request data in DB. Err: %v\nInsertManyResult: %v\n", err, result)
		return err
	}
	return err
}

func (m mongoDAO) UpdateFriendRequestsStatus(ctx context.Context, userId string, friendIds []string, status models.FriendRequestStatus) error {

	var allsuitableRecords []bson.M

	/*
		Example:
			db.friends.find(
				{
					$or:[
						{
							$or: [
								{userId:"112", friendId: "111"},
								{userId:"111", friendId: "112"}
							]
						},
						{
							$or: [
								{userId:"111", friendId: "113"},
								{userId:"113", friendId: "111"}
							]
						}
						]
				});
	*/

	for _, friendId := range friendIds {
		filterData := bson.M{
			literals.MongoOr: []bson.M{
				{
					literals.MongoUserId:   userId,
					literals.MongoFriendId: friendId,
				},
				{
					literals.MongoUserId:   friendId,
					literals.MongoFriendId: userId,
				},
			},
		}
		allsuitableRecords = append(allsuitableRecords, filterData)
	}

	filter := bson.M{
		literals.MongoOr: allsuitableRecords,
	}

	update := bson.M{
		literals.MongoSet: bson.M{
			literals.MongoStatus: status,
		},
	}

	result, err := m.databse.Collection(literals.FriendsCollection).UpdateMany(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to update friend request status in DB. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return err
}

func (m mongoDAO) RemoveFriends(ctx context.Context, userId string, friendIds []string) error {

	var allsuitableRecords []bson.M

	for _, friendId := range friendIds {
		filterData := bson.M{
			literals.MongoOr: []bson.M{
				{
					literals.MongoUserId:   userId,
					literals.MongoFriendId: friendId,
				},
				{
					literals.MongoUserId:   friendId,
					literals.MongoFriendId: userId,
				},
			},
		}
		allsuitableRecords = append(allsuitableRecords, filterData)
	}

	filter := bson.M{
		literals.MongoOr: allsuitableRecords,
	}

	result, err := m.databse.Collection(literals.FriendsCollection).DeleteMany(ctx, filter)
	if err != nil {
		fmt.Printf("Failed to update friend request status in DB. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return err
}

func (m mongoDAO) UpdateGamePartyStatus(ctx context.Context, partyIds []string, status models.GamePartyStatus) error {

	filter := bson.M{
		literals.MongoID: bson.M{literals.MongoIn: partyIds},
	}

	update := bson.M{
		literals.MongoSet: bson.M{
			literals.MongoStatus: status,
		},
	}

	result, err := m.databse.Collection(literals.GamePartyCollection).UpdateMany(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to update game party status in DB. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return nil
}

func (m mongoDAO) FetchActiveGameParties(ctx context.Context) ([]*models.GameParty, error) {

	filter := bson.M{
		literals.MongoStatus: models.GamePartyStatusActive,
	}
	/*
		// more safer query
			filter := bson.D{
				{literals.MongoStatus, models.GamePartyStatusStatusActive},
				// literals.Mongo
				// time since start time + duration < curr time
				{"$expr", bson.D{
					{"$gt", bson.A{
						bson.D{
							{"$add", bson.A{"$startTime", "$duration"}},
						},
						time.Now(),
					}},
				}},
			}
	*/

	cur, err := m.databse.Collection(literals.GamePartyCollection).Find(ctx, filter)
	if err != nil {
		fmt.Println("Error occurred while calling gameparty collection.", err)
		return nil, err
	}

	var gameParties []*models.GameParty
	for cur.Next(ctx) {
		var gameParty models.GameParty
		decodeErr := cur.Decode(&gameParty)
		if decodeErr != nil {
			fmt.Println("Failed to decode game party document.", decodeErr)
			return nil, decodeErr
		}
		gameParties = append(gameParties, &gameParty)
	}

	if len(gameParties) == 0 {
		fmt.Println("No active game parties found")
		return nil, nil
	}

	return gameParties, nil
}

func (m mongoDAO) CreateGameParty(ctx context.Context, gameParty *models.GameParty) error {

	docs := bson.M{
		literals.MongoID:        gameParty.PartyId,
		literals.MongoCreatedBy: gameParty.CreatedBy,
		literals.MongoStartTime: gameParty.StartTime,
		literals.MongoDuration:  gameParty.Duration,
		literals.MongoStatus:    gameParty.Status,
	}

	result, err := m.databse.Collection(literals.GamePartyCollection).InsertOne(ctx, docs)
	if err != nil {
		fmt.Printf("failed to insert game party in DB. Err: %v\nInsertOneResult: %v\n", err, result)
		return err
	}
	return nil
}

// check if all friendsIDs are friends of the user
func (m mongoDAO) CheckFriendship(ctx context.Context, userId string, friendIds []string) (bool, error) {

	filter := bson.M{
		literals.MongoUserId:   userId,
		literals.MongoStatus:   models.FriendshipStatusAccepted,
		literals.MongoFriendId: bson.M{literals.MongoIn: friendIds},
	}

	cur, err := m.databse.Collection(literals.FriendsCollection).Find(ctx, filter)
	if err != nil {
		fmt.Println("Error occurred while calling friends. ", err)
		return false, err
	}

	var friends []*models.Friends
	for cur.Next(ctx) {
		var friend models.Friends
		decodeErr := cur.Decode(&friend)
		if decodeErr != nil {
			fmt.Println(decodeErr)
			return false, decodeErr
		}
		friends = append(friends, &friend)
	}

	if len(friends) == 0 {
		fmt.Println("No friends found")
		return false, errors.New("no friends found")
	} else if len(friends) != len(friendIds) {
		logrus.WithFields(logrus.Fields{
			literals.LLRequestedFriendIds: friendIds,
			literals.LLFriendsFound:       (fmt.Sprintf("%+v", friends)),
		}).Error("requested users are not friends")
		return false, errors.New("requested users are not friends")
	}

	return true, nil
}

func (m mongoDAO) AddInviteesToGamePartyCollection(ctx context.Context, partyId string, newInvitees []string) error {

	var updates bson.M = bson.M{}

	for _, invitee := range newInvitees {
		updates[literals.MongoPlayersDotAccess+invitee] = models.PlayerInvitedStatus
	}

	filter := bson.M{
		literals.MongoID: partyId,
	}

	update := bson.M{
		literals.MongoSet: updates,
	}

	result, err := m.databse.Collection(literals.GamePartyCollection).UpdateMany(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to add invitees to the game party in the DB. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return nil
}

func (m mongoDAO) UpdatePlayersDecisionForGameParty(ctx context.Context, partyId string, userIds []string, playerStatus models.GamePartyPlayerStatus) error {

	var updates bson.M = bson.M{}
	var userFilters []bson.M = []bson.M{}

	for _, userId := range userIds {
		updates[literals.MongoPlayersDotAccess+userId] = playerStatus

		userFilters = append(userFilters, bson.M{literals.MongoPlayersDotAccess + userId: bson.M{literals.MongoExists: true}})
	}

	filter := bson.M{
		literals.MongoID:  partyId,
		literals.MongoAnd: userFilters,
	}

	update := bson.M{
		literals.MongoSet: updates,
	}

	result, err := m.databse.Collection(literals.GamePartyCollection).UpdateMany(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to update players status in the game party collection. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return nil
}

/*
// In order to use transactions, you need a MongoDB replica set,
func (m mongoDAO) UpdatePlayerAndUserStatusForGameParty(ctx context.Context, partyId string, userId string, playerStatus models.GamePartyPlayerStatus, userStatus models.UserStatus) error {

	filter := bson.M{
		literals.MongoID:               partyId,
		literals.MongoPlayers + userId: bson.M{literals.MongoExists: true},
	}

	update := bson.M{
		literals.MongoSet: bson.M{
			literals.MongoPlayers + userId: playerStatus,
		},
	}

	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc)

	session, err := m.client.StartSession()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			literals.LLInternalError: err,
			literals.LLPartyId:       partyId,
			literals.LLUserId:        userId,
		}).Error("failed to start a mongo session")
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		result, err := m.databse.Collection(literals.GamePartyCollection).UpdateOne(ctx, filter, update)
		if err != nil {
			fmt.Printf("Failed to update player status in the game party collection. Err: %v\nUpdateResult: %v\n", err, result)
			return nil, err
		}

		// once player has joined a game party, update his status to "in-game"
		result, err = m.UpdateUserStatus(ctx, userId, userStatus)
		if err != nil {
			return nil, err
		}
		return result, nil
	}, txnOptions)

	return err
}
*/

/*
// was created to append and remove data to/from arrays
// pull user from one array and push to another
func (m mongoDAO) PullAndPushDataInGamePartyCollection(ctx context.Context, partyId string, userId string, removeFromField string, addToField string) error {

	filter := bson.M{
		literals.MongoID: partyId,
	}

	update := bson.M{
		literals.MongoPull: bson.M{
			removeFromField: userId, //literals.MongoGamePartyInvitees
		},
		literals.MongoPush: bson.M{
			addToField: userId, //literals.MongoGamePartyAccepted
		},
	}

	result, err := m.databse.Collection(literals.GamePartyCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Printf("Failed to pull and push user in the game party collection. Err: %v\nUpdateResult: %v\n", err, result)
		return err
	}
	return nil
}
*/

/*
// fetch game parties that should have ended.// will have (start time + duration) < curr time
func (m mongoDAO) FetchGamePartiesToBeEnded(ctx context.Context) ([]*models.GameParty, error) {
	filter := bson.D{
		{literals.MongoStatus, models.GamePartyStatusActive},

		{"$expr", bson.D{
			{"$lt", bson.A{
				bson.D{
					{"$add", bson.A{"$startTime", "$duration"}},
				},
				time.Now(),
			}},
		}},
	}

	cur, err := m.databse.Collection(literals.GamePartyCollection).Find(ctx, filter)
	if err != nil {
		fmt.Println("Error occurred while calling gameparty collection.", err)
		return nil, err
	}

	var gameParties []*models.GameParty
	for cur.Next(ctx) {
		var gameParty models.GameParty
		decodeErr := cur.Decode(&gameParty)
		if decodeErr != nil {
			fmt.Println("Failed to decode game party document.", decodeErr)
			return nil, decodeErr
		}
		gameParties = append(gameParties, &gameParty)
	}

	if len(gameParties) == 0 {
		fmt.Println("No active game parties that should have ended")
		return nil, nil
	}

	return gameParties, nil
}
*/
