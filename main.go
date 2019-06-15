package main

import (
	"github.com/caarlos0/env/v5"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"

	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type (
	Config struct {
		DbHost               string `env:"DB_HOST" envDefault:"127.0.0.1"`
		DbPort               string `env:"DB_PORT" envDefault:"27017"`
		DbName               string `env:"DB_NAME" envDefault:"spyfall"`
		DbGameCollection     string `env:"DB_GAME_COLLETION" envDefault:"games"`
		DbPlayerCollection   string `env:"DB_PLAYER_COLLECTION" envDefault:"players"`
		DbLocationCollection string `env:"DB_LOC_COLLECTION" envDefault:"locations"`
		HTTPAddr             string `env:"HTTP_ADDRESS" envDefault:"0.0.0.0"`
		HTTPPort             string `env:"HTTP_PORT" envDefault:"80"`
		HTTPDir              string `env:"HTTP_DIR" envDefault:"public"`
	}

	UserConnection struct {
		Player PlayerEntry
		Game   GameEntry
	}
)

var (
	dbConn DatabaseConnection

	errorOther            = "OTHER_ERROR"
	errorGameExists       = "GAME_EXISTS_ERROR"
	errorGameNotFound     = "GAME_NOT_FOUND_ERROR"
	errorGameNotRemoved   = "GAME_NOT_REMOVED_ERROR"
	errorGameInProgress   = "GAME_IN_PROGRESS"
	errorUsernameExists   = "PLAYER_EXISTS_ERROR"
	errorPlayerNotFound   = "PLAYER_NOT_FOUND_ERROR"
	errorPlayerNotRemoved = "PLAYER_NOT_REMOVED_ERROR"
	errorNoGameCode       = "NO_GAME_CODE_ERROR"
	errorNoUsername       = "NO_USERNAME_ERROR"
	errorCodeExists       = "GAME_EXISTS_ERROR"
	errorInvalidUserName  = "INVALID_USER_NAME_ERROR"

	//
	connectedPlayers = make(map[bson.ObjectId][]*websocket.Conn) //Map the game's id to an array of websocket connections
	players          = make(map[*websocket.Conn]bson.ObjectId)   //Map the player's connection to their id
	connections      = make(map[bson.ObjectId]*websocket.Conn)   //Map the player's id to their connection
	//
)

/*
TODO:
 - When someone creates a game, they are not  added to the players collection
 - When the last player leaves a game, the game is not deleted

 - Test getGame to see what it returns when no game exists
 - Add handler for disconnections
*/

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
	}

	fmt.Println("Attempting to connect to database at: " + cfg.DbHost + ":" + cfg.DbPort)
	var err error
	dbConn, err = dbConnect(DbOptions{
		Server:             cfg.DbHost + ":" + cfg.DbPort,
		Database:           cfg.DbName,
		GameCollection:     cfg.DbGameCollection,
		PlayerCollection:   cfg.DbPlayerCollection,
		LocationCollection: cfg.DbLocationCollection,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("Unable to contact database. Shutting down.")
	}
	fmt.Println("Connected Sucessfully")

	http.Handle("/", http.FileServer(http.Dir(cfg.HTTPDir)))
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		s, err := newSocketRouter(w, r)
		if err != nil {
			fmt.Println(err)
		}

		s.addRoute("createGame", createGame)

		s.addRoute("destroyGame", destroyGame)

		s.addRoute("joinGame", joinGame)

		s.addRoute("leaveGame", leaveGame)

		s.addRoute("startGame", startGame)

		s.addRoute("stopGame", stopGame)

		s.addDisconnect(func(ctx SocketContext) {
		})

		s.handleRoutes()
	})
	http.ListenAndServe(":"+cfg.HTTPPort, nil)
}

func createGame(ctx SocketContext) SocketResponse {
	username := ctx.Data["username"].(string)

	responseData := ResponseData{
		Sucess:   true,
		Username: username,
	}

	if username == "" {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorInvalidUserName,
			Desc: "The username: \"" + username + "\" is invalid.",
		}

		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	game, err := dbConn.addGame()
	if err != nil {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorOther,
			Desc: err.Error(),
		}

		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	ctx.Prop["code"] = game.Code
	return joinGame(ctx)
}

func destroyGame(ctx SocketContext) SocketResponse {
	responseData := ResponseData{Sucess: true}
	player, err := dbConn.getPlayer(players[ctx.Connection])
	if err != nil {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorOther,
			Desc: err.Error(),
		}
		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	responseData.Code = gameCodes[player.Game]
	responseData.Username = player.Username

	err = dbConn.delGame(player.Game)
	if err != nil {
		if err.Error() == errorGameNotFound {
			responseData.Error = ResponseError{
				Code: err.Error(),
				Desc: "Game: " + responseData.Code + " not found in the database.",
			}
		} else {
			responseData.Error = ResponseError{
				Code: errorOther,
				Desc: err.Error(),
			}
		}
		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	for _, ctx := range connectedPlayers[player.Game] {
		ctx.WriteJSON()
	}

	delete(connectedPlayers, player.Game) //TODO: Add a function here to inform the connected sockets that the game is removed

	return leaveGame(ctx)
}

func joinGame(ctx SocketContext) SocketResponse {
	code, ok := ctx.Prop["code"].(string)
	if !ok {
		code = ctx.Data["code"].(string)
	}

	username := ctx.Data["username"].(string)
	responseData := ResponseData{
		Sucess:   true,
		Code:     code,
		Username: username,
	}
	
	gid, ok := gameIds[code]
	if !ok {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorGameNotFound,
			Desc: "Game '" + code + "' was not found. Is the game code correct?",
		}
	}

	player, err := dbConn.addPlayer(username, gid)
	if err != nil {
		responseData.Sucess = false

		switch err.Error() {
		case errorGameInProgress:
			responseData.Error = ResponseError{
				Code: err.Error(),
				Desc: "Cannot join a game that is currently in progress.",
			}
		case errorUsernameExists:
			responseData.Error = ResponseError{
				Code: err.Error(),
				Desc: "Player with the username: " + username + " already exists in game: " + code + ".",
			}
		default:
			responseData.Error = ResponseError{
				Code: errorOther,
				Desc: err.Error(),
			}
		}
		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	connectedPlayers[gid] = append(connectedPlayers[gid], ctx.Connection)
	connections[player.ID] = ctx.Connection
	players[ctx.Connection] = player.ID

	return SocketResponse{
		Type: ctx.Type,
		Data: responseData,
	}
}

func leaveGame(ctx SocketContext) SocketResponse {
	responseData := ResponseData{Sucess: true}
	player, err := dbConn.getPlayer(players[ctx.Connection])
	if err != nil {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorOther,
			Desc: err.Error(),
		}
		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	responseData.Code = gameCodes[player.Game]
	responseData.Username = player.Username

	err = dbConn.delPlayer(player.ID)
	if err != nil {
		responseData.Sucess = false
		responseData.Error = ResponseError{
			Code: errorOther,
			Desc: err.Error(),
		}

		return SocketResponse{
			Type: ctx.Type,
			Data: responseData,
		}
	}

	delete(connections, player.ID) //TODO: Add a function here to inform the connected websockets which are in this game that a player disconnected
	delete(players, ctx.Connection)

	return SocketResponse{
		Type: ctx.Type,
		Data: responseData,
	}
}

func startGame(ctx SocketContext) SocketResponse {
	return SocketResponse{}
}

func stopGame(ctx SocketContext) SocketResponse {
	return SocketResponse{}
}

func generateCode() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyz")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 6)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func notNil(test interface{}) interface{} {
	if test != nil {
		return test
	}
	return ""
}
