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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDAO interface {
	GetFreinds(ctx context.Context, userId string) ([]*models.User, error)
	GetUserDetails(ctx context.Context, userIds []string) ([]*models.User, error)
	StoreFriendRequests(ctx context.Context, userId string, friendIds []string) error
	UpdateFriendRequestsStatus(ctx context.Context, userId string, friendIds []string, status models.FriendRequestStatus) error
	RemoveFriends(ctx context.Context, userId string, friendIds []string) error

	// game party
	FetchActiveGameParties(ctx context.Context) ([]*models.GameParty, error)
	CreateGameParty(ctx context.Context, gamePary *models.GameParty) error
	UpdateGamePartyStatus(ctx context.Context, partyIds []string, status models.GamePartyStatus) error
}

var mongoDAOStruct MongoDAO
var mongodaoOnce sync.Once

type mongoDAO struct {
	databse *mongo.Database
	// mongodao

}

func InitMongoDao(db *mongo.Database) MongoDAO {
	mongodaoOnce.Do(func() {
		mongoDAOStruct = &mongoDAO{
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
func (m mongoDAO) GetFreinds(ctx context.Context, userId string) ([]*models.User, error) {

	filter := bson.M{
		literals.MongoUserId: userId,
		literals.MongoStatus: models.StatusAccepted,
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
		fmt.Printf("Some users not found, \nRequestedUserIds: %v\nFoundUsers: \n%v\n", userIds, users)
		return nil, errors.New("some users not found")
	}
	return users, err
}

func (m mongoDAO) StoreFriendRequests(ctx context.Context, userId string, friendIds []string) error {

	var docs []interface{}

	time := time.Now()

	for _, friendId := range friendIds {
		docs = append(docs, bson.M{
			literals.MongoID:          uuid.NewString(),
			literals.MongoUserId:      userId,
			literals.MongoStatus:      models.StatusPending,
			literals.MongoRequestedBy: userId,
			literals.MongoFriendId:    friendId,
			literals.MongoRequestedOn: time,
		})

		docs = append(docs, bson.M{
			literals.MongoID:          uuid.NewString(),
			literals.MongoUserId:      friendId,
			literals.MongoStatus:      models.StatusPending,
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

// this should be called at service startup
// fetch game parties that should have ended.
// will have (start time + duration) < curr time
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
