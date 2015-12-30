package thordb

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"thorium-go/globals"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

const privKeyPath string = "keys/app.rsa"
const pubKeyPath string = "keys/app.rsa.pub"

// redis keys
const sessionKey string = "sessions/user/%d"
const hkeyUserToken string = "userToken"
const hkeyCharacterToken string = "characterToken"
const hkeyCharacterData string = "characterData"
const gameSessionKey string = "games/%d"

var db *sql.DB
var kvstore *redis.Client
var signKey *rsa.PrivateKey
var verifyKey *rsa.PublicKey

func init() {
	// check rsa
	var signBytes []byte
	var verifyBytes []byte
	var err error
	log.Print("opening app.rsa keys")
	signBytes, err = ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Print(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Print(err)
	}
	verifyBytes, err = ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Print(err)
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Print(err)
	}

	log.Print("testing postgres connection")
	// check postgres
	db, err = sql.Open("postgres", "port=5432 host=db user=postgres password=secret dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Print(err)
	}

	log.Print("testing redis connection")
	// check redis
	kvstore = redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})

	_, err = kvstore.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("thordb initialization complete")
}

func RegisterNewGame(mapName string, maxPlayers int) (int, error) {
	var gameId int
	err := db.QueryRow("INSERT INTO games (map_name, max_players) VALUES ( $1, $2 ) RETURNING game_id", mapName, maxPlayers).Scan(&gameId)
	return gameId, err
}

func RegisterActiveGame(game_id int, machine_id int, port int) (bool, error) {

	res, err := db.Exec("INSERT INTO game_servers (game_id, machine_id, port) VALUES ($1, $2, $3)", game_id, machine_id, port)
	if err != nil {
		return false, err
	}

	var rows int64
	rows, err = res.RowsAffected()
	if err != nil {
		return false, err
	}

	exists := rows > 0
	return exists, err
}

func RegisterAccount(username string, password string) (string, []int, error) {

	var foundname string

	// check to see if username is taken already
	err := db.QueryRow("SELECT username FROM account_data WHERE username LIKE $1;", username).Scan(&foundname)
	switch {
	case err == sql.ErrNoRows:
		log.Print("Username available")
	case err != nil:
		log.Print(err)
		return "", nil, err
	default:
		log.Print("Username is already in use")
		return "", nil, errors.New("thordb: already in use")
	}

	saltSize := 16
	alg := "sha1"

	//allocates 16+sha1.Size bytes to the bufer
	//creates slice with length saltSize and capacity of saltSize+sha1.Size
	buf := make([]byte, saltSize, saltSize+sha1.Size)

	//fill buf with random data (linux is /dev/urandom)
	_, e := io.ReadFull(rand.Reader, buf)

	if e != nil {
		fmt.Println("filling buf with random data failed")
		return "", nil, e
	}

	// create password hash
	dirtySalt := sha1.New()
	dirtySalt.Write(buf)
	dirtySalt.Write([]byte(password))
	salt := dirtySalt.Sum(buf)
	combination := string(salt) + string(password)
	passwordHash := sha1.New()
	io.WriteString(passwordHash, combination)

	var uid int
	timenow := time.Now()

	// register new account in the database
	err = db.QueryRow("INSERT INTO account_data (username, password, salt, algorithm, createdon, lastlogin) VALUES ($1, $2, $3, $4, $5, $6) RETURNING user_id", username, passwordHash.Sum(nil), salt, alg, timenow, timenow).Scan(&uid)
	if err != nil {
		fmt.Println("error inserting account data: ", err)
		return "", nil, err
	}

	// create the jwt token data
	t := jwt.New(jwt.SigningMethodRS256)
	t.Claims["uid"] = uid
	t.Claims["iat"] = time.Now()

	// create signed token string
	token, err := t.SignedString(signKey)
	if err != nil {
		return "", nil, err
	}

	// grab the character ids from db
	// this should always be empty but check anyway
	var charIds []int
	rows, err := db.Query("SELECT id FROM characters where uid=$1", uid)
	if err != nil {
		log.Print("error querying character ids from uid: ", err)
		return "", nil, err
	}
	defer rows.Close()
	var charId int
	for rows.Next() {
		err = rows.Scan(&charId)
		if err != nil {
			log.Print("error scanning row to get character ID: ", err)
		}
		charIds = append(charIds, charId)
	}

	// set the session in redis and give it an expiry
	key := fmt.Sprintf(sessionKey, uid)
	kvstore.HSet(key, hkeyUserToken, token)
	kvstore.Expire(key, time.Second*globals.SESSION_EXPIRE_SECONDS)

	return token, charIds, nil
}

func LoginAccount(username string, password string) (string, []int, error) {

	var hashedPassword []byte
	var salt []byte
	var uid int

	// get the account info from the database
	err := db.QueryRow("SELECT password, salt, user_id FROM account_data WHERE username LIKE $1", username).Scan(&hashedPassword, &salt, &uid)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("thordb: user does not exist %s", username)
		return "", nil, errors.New("thordb: does not exist")
	case err != nil:
		log.Print(err)
		return "", nil, err
	}

	combination := string(salt) + string(password)
	passwordHash := sha1.New()
	io.WriteString(passwordHash, combination)

	// compare password hashes
	match := bytes.Equal(passwordHash.Sum(nil), hashedPassword)
	if !match {
		return "", nil, errors.New("thordb: invalid password")
	}

	// create the jwt token data
	t := jwt.New(jwt.SigningMethodRS256)
	t.Claims["uid"] = uid
	t.Claims["iat"] = time.Now()

	// create signed token string
	token, err := t.SignedString(signKey)
	if err != nil {
		return "", nil, err
	}

	// first check if a session already exists, if so reject as "already logged on" unless the time is substantially old (> 5min)
	var alreadyLoggedIn bool = true

	_, err = kvstore.HGet(fmt.Sprintf(sessionKey, uid), hkeyUserToken).Result()
	if err != nil {
		switch err.Error() {
		case "redis: nil":
			alreadyLoggedIn = false
		default:
			return "", nil, err
		}
	}

	if alreadyLoggedIn {
		return "", nil, errors.New("thordb: already logged in")
	}

	//grab the character ids from db
	var charIds []int
	rows, err := db.Query("SELECT id FROM characters where uid=$1", uid)
	if err != nil {
		log.Print("error querying character ids from uid: ", err)
		return "", nil, err
	}
	defer rows.Close()
	var charId int
	for rows.Next() {
		err = rows.Scan(&charId)
		if err != nil {
			log.Print("error scanning row to get character ID: ", err)
		}
		charIds = append(charIds, charId)
	}

	// set the session in redis and give it an expiry
	key := fmt.Sprintf(sessionKey, uid)
	kvstore.HSet(key, hkeyUserToken, token)
	kvstore.Expire(key, time.Second*globals.SESSION_EXPIRE_SECONDS)

	return token, charIds, nil
}

func Disconnect(userToken string) error {

	uid, err := validateToken(userToken)
	if err != nil {
		return err
	}

	var charToken string
	var charData string
	var foundCharacter bool = true

	charToken, err = kvstore.HGet(fmt.Sprintf(sessionKey, uid), hkeyCharacterToken).Result()
	if err != nil {
		// no character to save
		switch err.Error() {
		case "redis: nil":
			// no character to save
			foundCharacter = false
		default:
			return err
		}
		log.Print(err)
	}

	// decrypt the token and get character id
	if foundCharacter {
		var token *jwt.Token
		token, err = jwt.Parse(charToken, func(token *jwt.Token) (interface{}, error) {
			return verifyKey, nil
		})
		if err != nil {
			log.Print("thordb couldn't parse stored character token")
			log.Print(err)
			return err
		}
		idFloat, ok := token.Claims["id"].(float64)
		if !ok {
			log.Print("thordb couldn't parse stored character token")
			log.Print(err)

		}
		id := int(idFloat)
		charData, err = kvstore.HGet(fmt.Sprintf(sessionKey, uid), hkeyCharacterData).Result()
		if err != nil {
			// no character to save
			switch err.Error() {
			case "redis: nil":
				// no character to save
			default:
				return err
			}
		}

		var res sql.Result
		res, err = db.Exec("UPDATE characters SET game_data = $1 WHERE id = $2 AND uid = $3", charData, id, uid)
		if err != nil {
			return err
		}

		var rows int64
		rows, err = res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			// character does not exist
			return errors.New("thordb: does not exist")
		}

		res, err = db.Exec("UPDATE account_data SET lastlogin = $1 WHERE user_id = $2", time.Now(), uid)
		if err != nil {
			return err
		}

		rows, err = res.RowsAffected()
		if err != nil {
			return err
		}

		if rows == 0 {
			// character does not exist
			return errors.New("thordb: does not exist")
		}

		if err != nil {
			return err
		}

	}

	var count int64
	count, err = kvstore.Del(fmt.Sprintf(sessionKey, uid)).Result()
	if err != nil {
		return err
	}

	if count == 0 {
		log.Print("couldnt find session")
		return errors.New("thordb: invalid session")
	}

	log.Print("client disconnected %d", uid)
	return nil
}

// helper funcs
func storeAccount(session *AccountSession) {
	// use this to store an account update in postgres
}

func validateToken(token_str string) (int, error) {

	token, err := jwt.Parse(token_str, func(t *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	if err != nil {
		return 0, err
	}

	var uidFloat64 float64
	uidFloat64, ok := token.Claims["uid"].(float64)
	uid := int(uidFloat64)
	if !ok {
		return 0, errors.New("thordb: invalid session")
	}

	// ToDo: update account + character in postgres before deleting from redis

	var savedToken string
	savedToken, err = kvstore.HGet(fmt.Sprintf(sessionKey, uid), hkeyUserToken).Result()

	if err != nil {
		return 0, err
	}

	if token_str == savedToken {
		return uid, nil
	} else {
		return 0, errors.New("thordb: invalid session")
	}
}

func CreateCharacter(userToken string, character *CharacterData) (*CharacterSession, error) {

	uid, err := validateToken(userToken)
	if err != nil {
		return nil, err
	}

	var foundname string
	err = db.QueryRow("SELECT name FROM characters WHERE name LIKE $1", character.Name).Scan(&foundname)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("thordb: name is available %s", character.Name)
	case err != nil:
		log.Print(err)
		return nil, err
	default:
		return nil, errors.New("thordb: already in use")
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(character)
	if err != nil {
		return nil, err
	}

	var id int
	err = db.QueryRow("INSERT INTO characters (uid, name, game_data) VALUES ($1, $2, $3) RETURNING id", uid, character.Name, string(jsonBytes)).Scan(&id)
	if err != nil {
		return nil, err
	}

	// create new jwt token with character id in claims and return new session with it
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["uid"] = uid
	token.Claims["id"] = id
	token.Claims["iat"] = time.Now()

	var token_str string
	token_str, err = token.SignedString(signKey)
	if err != nil {
		return nil, err
	}

	charSession := NewCharacterSessionFrom(character)
	charSession.UserID = uid
	charSession.ID = id
	charSession.Token = token_str

	// search db for any existing tutorial games
	var game_id int
	err = db.QueryRow("SELECT game_id FROM games WHERE game_mode LIKE $1 ORDER BY RANDOM() LIMIT 1", "tutorial").Scan(&game_id)
	switch {

	case err == sql.ErrNoRows:
		log.Print("thordb: no tutorial games found")

		var game_id int
		err = db.QueryRow("INSERT INTO games (map_name, game_mode) VALUES ('Mode_Sandbox', 'tutorial') RETURNING game_id").Scan(&game_id)
		if err != nil {
			return nil, err
		}

		// todo
		// provision new game on an available machine here
		err = ProvisionNewGame(game_id, "Mode_Sandbox", "tutorial")
		if err != nil {
			return nil, err
		}

		// then
		// return character session on new tutorial game
		charSession.GameId = game_id
		return charSession, nil

	case err != nil:
		return nil, err

	}

	charSession.GameId = game_id
	return charSession, nil
}

func SelectCharacter(userToken string, id int) (*CharacterSession, error) {

	uid, err := validateToken(userToken)
	if err != nil {
		return nil, err
	}

	// read from characters table and get game_data json string
	var game_data string
	var game_id int
	err = db.QueryRow("SELECT game_data, last_game_id from characters WHERE uid = $1 AND id = $2", uid, id).Scan(&game_data, &game_id)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var character CharacterData
	json.Unmarshal([]byte(game_data), &character)
	if err != nil {
		log.Print("thordb: unable to construct character from db data")
		log.Print(err)
		return nil, err
	}

	// todo: create new jwt token with character id in claims and return new session with it
	// create the jwt token
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["uid"] = uid
	token.Claims["id"] = id
	token.Claims["iat"] = time.Now()

	var token_str string
	token_str, err = token.SignedString(signKey)
	if err != nil {
		return nil, err
	}

	charSession := NewCharacterSessionFrom(&character)
	charSession.UserID = uid
	charSession.ID = id
	charSession.Token = token_str

	// if most recent session is 0 then the player has never played a game
	// therefore we let them join an availabe tutorial instance or start a new one
	if game_id == 0 {
		err = db.QueryRow("SELECT game_id FROM games WHERE game_mode LIKE $1 ORDER BY RANDOM() LIMIT 1", "tutorial").Scan(&game_id)
		switch {
		case err == sql.ErrNoRows:
			log.Print("thordb: no tutorial games found")

			var game_id int
			err = db.QueryRow("INSERT INTO games (map_name, game_mode) VALUES ('Mode_Sandbox', 'tutorial') RETURNING game_id").Scan(&game_id)
			if err != nil {
				return nil, err
			}

			// todo
			// provision new game on an available machine here
			err = ProvisionNewGame(game_id, "Mode_Sandbox", "tutorial")
			if err != nil {
				return nil, err
			}

			// then
			// return character session on new tutorial game
			charSession.GameId = game_id

		case err != nil:
			return nil, err

		default:
			charSession.GameId = game_id
		}
	} else {
		// game_id was not zero, so try to return to last game session
		// if it doesnt exist as an active game, then reload it onto a new machine
		// todo: what if it is a completed game? probably return to a fallback location such as home base
		var map_name string
		var game_mode string
		err = db.QueryRow("SELECT (map_name, game_mode) FROM games WHERE game_id = $1", game_id).Scan(&map_name, &game_mode)
		// game never existed but this should be a major error on our part because we got this id from the db earlier
		// so lets not handle it gracefully
		if err != nil {
			return nil, err
		}

		var address string
		var port int
		err = db.QueryRow("SELECT (remote_address, port) FROM game_servers JOIN machines USING (machine_id) WHERE game_id = $1", game_id).Scan(&address, &port)
		switch {

		case err == sql.ErrNoRows:
			log.Print("thordb: game not found")

			err = ProvisionNewGame(game_id, map_name, game_mode)
			if err != nil {
				return nil, err
			}

			// then
			// return character session on new tutorial game
		}

		charSession.GameId = game_id
	}
	err = kvstore.HSet(fmt.Sprintf(sessionKey, charSession.UserID), hkeyCharacterToken, charSession.Token).Err()
	if err != nil {
		log.Print("thordb: kvstore unreachable")
		log.Print(err)
	}

	err = kvstore.HSet(fmt.Sprintf(sessionKey, charSession.UserID), "characterData", game_data).Err()
	if err != nil {
		log.Print("thordb: kvstore unreachable")
		log.Print(err)
	}

	return charSession, nil
}

func GetServerInfo(game_id int) (string, int, error) {

	var count int
	err := db.QueryRow("SELECT count(*) from games WHERE game_id = $1", game_id).Scan(&count)
	if err != nil {
		return "", 0, err
	}

	if count == 0 {
		// game doesnt exist
		return "", 0, errors.New("thordb: does not exist")
	}

	var (
		address string
		port    int
	)

	err = db.QueryRow("SELECT remote_address, port from game_servers JOIN machines USING (machine_id) WHERE game_id = $1", game_id).Scan(&address, &port)
	if err != nil {
		log.Print(err)
		return "", 0, errors.New("thordb: game not available yet")
	}

	return address, port, nil
}

// ToDo: remove this func from public, only exposed for testing
// this should be used internally to thordb only!
func StoreCharacterSnapshot(charSession *CharacterSession) (bool, error) {
	b, err := json.Marshal(charSession.CharacterData)
	if err != nil {
		return false, err
	}

	var res sql.Result
	res, err = db.Exec("UPDATE characters SET game_data = $1 WHERE id = $2 AND uid = $3", string(b), charSession.ID, charSession.UserID)
	if err != nil {
		return false, err
	}

	var rowsAffected int64
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return false, err
	}

	if rowsAffected == 0 {
		return false, errors.New("thordb: does not exist")
	}

	return true, nil
}
