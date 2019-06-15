package main

import (
	"fmt"
	"reflect"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	DbOptions struct {
		Server             string
		Database           string
		GameCollection     string
		PlayerCollection   string
		LocationCollection string
	}

	DatabaseConnection struct {
		DB                 *mgo.Database
		GameCollection     *mgo.Collection
		PlayerCollection   *mgo.Collection
		LocationCollection *mgo.Collection
	}

	PlayerEntry struct {
		ID       bson.ObjectId `bson:"_id" json:"id"`
		Game     bson.ObjectId `bson:"game" json:"game"`
		Username string        `bson:"username" json:"username"`
		Role     string        `bson:"role" json:"role"`
		Spy      bool          `bson:"spy" json:"spy"`
	}

	GameEntry struct {
		ID       bson.ObjectId   `bson:"_id" json:"id"`
		Code     string          `bson:"code" json:"code"`
		Location string          `bson:"location" json:"location"`
		Players  []bson.ObjectId `bson:"players" json:"players"`
		Active   bool            `bson:"active" json:"active"`
	}
)

var (
	gameIds = make(map[string]bson.ObjectId) //Map the game's code to it's id
	gameCodes = make(map[bson.ObjectId]string) //Map the game's id to it's code
)

/*
	All of these functions assume the supplied game id is valid
*/

func dbConnect(dbo DbOptions) (DatabaseConnection, error) {
	session, err := mgo.Dial(dbo.Server)
	if err != nil {
		return DatabaseConnection{}, err
	}

	db := session.DB(dbo.Database)

	return DatabaseConnection{
		DB:                 db,
		GameCollection:     db.C(dbo.GameCollection),
		PlayerCollection:   db.C(dbo.PlayerCollection),
		LocationCollection: db.C(dbo.LocationCollection),
	}, nil
}

//Adds a new GameEntry to the database. Returns the added GameEntry and an Error
//Returns the origional GameEntry. If no errors, will include the generated ObjectID
//Can return an ErrorGameExists, as well as undocumented errors
func (dbConn *DatabaseConnection) addGame() (GameEntry, error) {
	code := generateCode()
	good := false
	for !good {
		id := gameIds[code]
		if id.Valid() {
			code = generateCode()
			fmt.Println("Duplicate code generated! Generating again!")
		} else {
			good = true
		}
	}

	result := GameEntry{
		ID:   bson.NewObjectId(),
		Code: code,
	}

	err := dbConn.GameCollection.Insert(result)
	if err != nil {
		return result, err
	}

	gameIds[result.Code] = result.ID
	gameCodes[result.ID] = result.Code
	return result, nil
}

//Removes a GameEntry from the database.
//Can retun an ErrorGameNotFound, as well as undocumented errors
func (dbConn *DatabaseConnection) delGame(id bson.ObjectId) error {
	game, err := dbConn.getGame(id)
	if err != nil {
		return err
	}
	
	if len(game.Players) > 0 {
		for _, p := range game.Players {
			dbConn.PlayerCollection.RemoveId(p)
		}
	}
	
	err = dbConn.GameCollection.RemoveId(id)
	if err != nil {
		return err
	}

	delete(gameIds, game.Code)
	delete(gameCodes, game.ID)
	return nil
}

//Returns the GameEntry that corresponds to the supplied GameID
//Returns the found GameEntry. Will be empty if there is an error
//Can return undocumented errors
func (dbConn *DatabaseConnection) getGame(id bson.ObjectId) (GameEntry, error) {
	result := GameEntry{}
	err := dbConn.GameCollection.FindId(id).One(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}

//Updates the specified GameEntry (Will not update code or ID). Rethrns the updated GameEntry and an Error
//GameEntry.Code is the only required field, but really that wouldn't make much sense now would it?
//Returns the origional GameEntry
//Can an ErrorGameNotFound and return undocumented errors
func (dbConn *DatabaseConnection) updateGame(g GameEntry) (GameEntry, error) {
	game, err := dbConn.getGame(g.ID)
	if err != nil {
		return g, fmt.Errorf(errorGameNotFound)
	}

	if game.Location != g.Location {
		err := dbConn.GameCollection.UpdateId(g.ID, bson.M{"$set": bson.M{"location": g.Location}})
		if err != nil {
			return g, err
		}
	}
	if g.Active != g.Active {
		err := dbConn.GameCollection.UpdateId(g.ID, bson.M{"$set": bson.M{"active": g.Active}})
		if err != nil {
			return g, err
		}
	}
	if !reflect.DeepEqual(game.Players, g.Players) {
		err := dbConn.GameCollection.UpdateId(g.ID, bson.M{"$set": bson.M{"players": g.Players}})
		if err != nil {
			return g, err
		}
	}

	return g, nil
}

//Adds a new PlayerEntry to the specified game. Returns the added PlayerEntry and an Error
//Can return an ErrorGameInProgress, ErrorPlayerExists, as well as undocumented errors
func (dbConn *DatabaseConnection) addPlayer(username string, gameId bson.ObjectId) (PlayerEntry, error) {
	p := PlayerEntry{
		Username: username,
		ID:       bson.NewObjectId(),
		Game:     gameId,
		Role:     "Counter-Spy",
		Spy:      false,
	}

	game, err := dbConn.getGame(p.Game)
	if err != nil {
		return p, err
	}

	if game.Active {
		return p, fmt.Errorf(errorGameInProgress)
	}

	if err := dbConn.checkUser(username, game.Code); err != nil {
		return p, err
	}

	if err := dbConn.PlayerCollection.Insert(p); err != nil {
		return p, err
	}

	return p, dbConn.GameCollection.UpdateId(gameId, bson.M{"$push": bson.M{"players": p.ID}})
}

//Removes a PlayerEntry from the database.
//Can return undocumented errors
func (dbConn *DatabaseConnection) delPlayer(pid bson.ObjectId) error {
	p, err := dbConn.getPlayer(pid)
	if err != nil {
		return err
	}

	err = dbConn.GameCollection.UpdateId(p.Game, bson.M{"$pull": bson.M{"players": p.ID}})
	if err != nil {
		return err
	}

	game, err := dbConn.getGame(p.Game)
	if err != nil {
		return err
	}

	if len(game.Players) == 0 {
		err := dbConn.delGame(game.ID)
		if err != nil {
			return err
		}
	}
	return dbConn.PlayerCollection.RemoveId(p.ID)
}

//Returns the PlayerEntry that corresponds to the supplied PlayerEntry
//Returns the found PlayerEntry. Will be empty if there is an error
//Can return undocumented errors
func (dbConn *DatabaseConnection) getPlayer(pid bson.ObjectId) (PlayerEntry, error) {
	result := PlayerEntry{}
	err := dbConn.PlayerCollection.FindId(pid).One(&result)
	return result, err
}

//Updates the specified PlayerEntry (Will not update GameID or ID). Returns the updated Player and an Error
//PlayerEntry.ID is the only required field, but really that wouldn't make much sense now would it?
//Returns the origional PlayerEntry
//Cen return undocumented errors
func (dbConn *DatabaseConnection) updatePlayer(p PlayerEntry) (PlayerEntry, error) {
	player, err := dbConn.getPlayer(p.ID)
	if err != nil {
		return p, err
	}

	if player.Username != p.Username {
		err := dbConn.PlayerCollection.UpdateId(p.ID, bson.M{"$set": bson.M{"username": p.Username}})
		if err != nil {
			return p, err
		}
	}
	if player.Role != p.Role {
		err := dbConn.PlayerCollection.UpdateId(p.ID, bson.M{"$set": bson.M{"role": p.Role}})
		if err != nil {
			return p, err
		}
	}
	if player.Spy != p.Spy {
		err := dbConn.PlayerCollection.UpdateId(p.ID, bson.M{"$set": bson.M{"spy": p.Spy}})
		if err != nil {
			return p, err
		}
	}

	return p, nil
}

//Returns an error if the player exists, returns nothing if it doesn't.
//GameCode and PlayerEntry.Username are the only fields that must be populated
//Can return an ErrorPlayerExists, ErrorGameNotFound as well as undocumented errors
func (dbConn *DatabaseConnection) checkUser(username, code string) error {
	id := gameIds[code]
	if id.Valid() {
		playerCount, err := dbConn.PlayerCollection.Find(bson.M{"game": id, "username": username}).Limit(1).Count()
		if err != nil {
			return err //If there is a problem with the Collection.Find()
		}
		if playerCount == 0 {
			return nil
		}
		return fmt.Errorf(errorUsernameExists)
	} else {
		return fmt.Errorf(errorGameNotFound)
	}

}
