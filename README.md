# lite-social-presence-system
Backend services for a lite social-presence system game using Golang

<hr>

**MongoDB** <i>[docker container]</i> 
   - Database: social-presence-system
   - Collections:
      -  users
      -  friends
      -  gameparty

<h4>Friends REST APIs</h4>

1. **PATCH /game/friends/request**
   - Add Friends: Users can send friend requests to multiple users

2. **PATCH /game/friends/handle-request**
   - Accept/Reject Friend Requests: Users can accept or reject friend requests they receive
   - <i>Notes:
       - Friendship is mutual
       - Sending same status for all the friend IDs in the request data</i>

3. **DELETE /game/friends/remove**
   - Remove Friends: Users can remove other users from their friend list

5. **GET /game/friends**
   - View Friend List: Users can view their current list of friends

<h4>Game Party REST APIs</h4>

1. **POST /game/party/create**
   - Create Game Party: Users can create a short game party session
2. **PATCH /game/party/invite**
   - Invite to Game Party: Users can invite their friends to join their game party
3. **PATCH /game/party/handle**
   - Handle Game Party: Users can give his decision as accepted/rejected for a game party invitation
4. **PATCH /game/party/join**
   - Join Game Party: Users can join the accepted game party
5. **POST /game/party/exit**
   - Exit Game Party: Users can exit from a game party
6. **PATCH /game/party/remove**
   - Remove from Game Party: Party leader can remove players from the game party

<h4>Real time update services</h4>

1. **User gets a notification whenever a player joins the party**
2. **User Friend gets a notification whenever he logs in**

<h4> Docker Compose</h4>

Inside the `lite-social-presence-system` project directory, 
run `sudo docker-compose up -d --build` to start the services.
Monitor logs using command `docker logs lite-social-presence-system_myproject_1 -f`.

For calling the REST APIs, import `GameRestApisPostman.json` file in Postman

For gRPC, invoke
1. localhost:8083
   UserService/StreamUserStatusChange
2. localhost:8083
   UserService/StreamPlayerJoinedStatus

<h4>minikube</h4>
start the minikube: `minikube start --driver=docker`

apply the deployment using: `kubectl apply -f deployment.yaml`

apply the service using: `kubectl apply -f service.yaml`

**Problem: Getting `Exiting due to SVC_UNREACHABLE: service not available: no running pod for service my-app found` error and not able to access the service using K8S

