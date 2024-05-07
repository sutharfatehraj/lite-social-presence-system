# lite-social-presence-system
Backend services for a lite social-presence system online-game using Golang

REST APIs

PATCH "/game/send-friend-request"
- Add Friends: Users can send friend requests to multiple users.

PATCH "/game/handle-friend-request"
- Accept/Reject Friend Requests: Users can accept or reject friend requests they receive.
Note: Friendship is mutual.

Users will get same status for all the friend IDs because you can select and send different statuses with every friend ID from UI.

DELETE "/game/remove-freinds"
- Remove Friends: Users can remove other users from their friend list.

GET "/game/friends"
- View Friend List: Users can view their current list of friends.

MongoDB database social-presence-system

collections:
gameparty

