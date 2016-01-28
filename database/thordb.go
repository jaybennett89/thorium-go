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
	"thorium-go/client"
	"thorium-go/globals"
	"thorium-go/model"
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

// errors
var ErrInvalidSessionKey = errors.New("thordb: invalid session key")
var ErrInvalidMachineKey = errors.New("thordb: invalid machine key")
var ErrGameNotExist = errors.New("thordb: game does not exist")
var ErrGameFull = errors.New("thordb: game is full")

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

func CreateNewGame(mapName string, gameMode string, minimumLevel int, maxPlayers int) (int, error) {

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}

	var gameId int
	err = tx.QueryRow("INSERT INTO games (map_name, game_mode, minimum_level, maximum_players) VALUES ( $1, $2, $3, $4 ) RETURNING game_id", mapName, gameMode, minimumLevel, maxPlayers).Scan(&gameId)
	if err != nil {
		return 0, err
	}

	machineList, err := GetMachineList()
	if err != nil {

		err = tx.Rollback()

		if err != nil {

			fmt.Println(err)
			return 0, err
		}

		fmt.Println(err)
		return 0, err
	}

	if len(machineList) < 1 {

		err = tx.Rollback()
		if err != nil {

			return 0, err
		}

		return 0, errors.New("thordb: no available servers")
	}

	// pick a machine
	// for now its okay to use the first one since we only have one server in dev environment
	machine := machineList[0]
	fmt.Println("selected ", machine.RemoteAddress, ":", machine.ListenPort)

	endpoint := fmt.Sprintf("%s:%d", machine.RemoteAddress, machine.ListenPort)
	rc, body, err := client.NewGameServer(endpoint, gameId, mapName, gameMode, minimumLevel, maxPlayers)
	if err != nil {

		err = tx.Rollback()
		if err != nil {

			return 0, err
		}

		return 0, err
	}

	fmt.Println("new game server response status : ", rc, " body : ", body)

	if rc != 200 {

		err = tx.Rollback()
		if err != nil {

			return 0, err
		}

		return 0, errors.New("thordb: machine unavailable")
	}

	_, err = tx.Exec("INSERT INTO loading_hosts (game_id, machine_id, kickoff_time) VALUES ( $1, $2, $3 )", gameId, machine.MachineId, time.Now())
	if err != nil {

		fmt.Println(err)

		err = tx.Rollback()
		if err != nil {

			return 0, err
		}

		return 0, err
	}

	err = tx.Commit()
	if err != nil {

		return 0, err
	}

	return gameId, nil
}

func RegisterActiveGame(gameId int, machineKey string, listenPort int) error {

	machineId, err := readMachineKey(machineKey)
	if err != nil {

		return err
	}

	tx, err := db.Begin()
	if err != nil {

		return err
	}

	_, err = tx.Exec("DELETE FROM loading_hosts WHERE game_id = $1 AND machine_id = $2", gameId, machineId)
	if err != nil {

		return err
	}

	_, err = tx.Exec("INSERT INTO hosts (game_id, machine_id, port) VALUES ( $1, $2, $3 )", gameId, machineId, listenPort)
	if err != nil {

		err = tx.Rollback()
		if err != nil {

			return err
		}

		return err
	}

	err = tx.Commit()
	if err != nil {

		return err
	}

	return nil
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
	var charIds []int = []int{}
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
	var charIds []int = []int{}
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

func readMachineKey(machineKey string) (machineId int, err error) {

	token, err := jwt.Parse(machineKey, func(t *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	if err != nil {

		return 0, err
	}

	var rawId float64
	rawId, ok := token.Claims["machineId"].(float64)
	if !ok {

		return 0, errors.New("thordb: invalid machine key")
	}

	machineId = int(rawId)

	return machineId, nil
}

func CreateCharacter(sessionKey string, name string, classId int) (int, error) {

	uid, err := validateToken(sessionKey)
	if err != nil {
		return 0, err
	}

	var foundname string
	err = db.QueryRow("SELECT name FROM characters WHERE name LIKE $1", name).Scan(&foundname)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("thordb: name is available %s", name)
	case err != nil:
		log.Print(err)
		return 0, err
	default:
		return 0, errors.New("thordb: already in use")
	}

	character := model.NewCharacter()
	character.Name = name
	character.SetClassAttributes(classId)

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(character.CharacterState)
	if err != nil {
		return 0, err
	}

	var id int
	err = db.QueryRow("INSERT INTO characters (uid, name, game_data) VALUES ($1, $2, $3) RETURNING id", uid, character.Name, string(jsonBytes)).Scan(&id)
	if err != nil {
		return 0, err
	}

	// need to store in cache?

	return id, nil
}

func GetServerInfo(gameId int) (*model.HostServer, bool, error) {

	// return model, true, nil if game exists and server is registered
	// return nil, false, nil if game exists but server is not loaded yet
	// return err otherwise

	var host model.HostServer

	err := db.QueryRow("SELECT remote_address, port FROM games JOIN hosts USING (game_id) JOIN machines USING (machine_id) WHERE game_id = $1", gameId).Scan(&host.RemoteAddress, &host.ListenPort)
	switch {

	// if game is not found in hosts then check loading_hosts too
	case err == sql.ErrNoRows:

		var kickoff time.Time
		err := db.QueryRow("SELECT kickoff_time FROM loading_hosts WHERE game_id = $1", gameId).Scan(&kickoff)
		switch {

		case err == sql.ErrNoRows:

			return nil, false, GameNotExistError

		case err != nil:

			return nil, false, err

		}

		// todo: compare kickoff time to current time and reprovision if elapsed time
		// is greater than a wait threshold constant

		return nil, false, nil

	case err != nil:

		return nil, false, err
	}

	host.GameId = gameId
	return &host, true, nil
}

func SelectCharacter(sessionKey string, characterId int) (*model.Character, error) {

	uid, err := validateToken(sessionKey)
	if err != nil {
		return nil, err
	}

	var character model.Character
	character.CharacterId = characterId

	var gameData string

	err = db.QueryRow("SELECT name, last_game_id, game_data FROM characters WHERE id = $1 AND uid = $2", characterId, uid).Scan(&character.Name, &character.LastGameId, &gameData)
	if err != nil {
		return nil, err
	}

	var state model.CharacterState
	err = json.Unmarshal([]byte(gameData), &state)
	if err != nil {
		return nil, err
	}

	character.CharacterState = state

	return &character, nil
}

func PlayerConnect(gameId int, machineKey string, sessionKey string, characterId int) (*model.Character, error) {

	machineId, valid, err := validateMachineKey(machineKey)
	if err != nil {

		return nil, err
	}

	if !valid {

		return nil, ErrInvalidMachineKey
	}

	userId, err := validateToken(sessionKey)
	if err != nil {

		return nil, ErrInvalidSessionKey
	}

	var playerCount int
	var maxPlayers int

	err = db.QueryRow("SELECT player_count, maximum_players FROM games JOIN hosts USING (game_id) JOIN machines USING (machine_id) WHERE game_id = $1 AND machine_id = $2", gameId, machineId).Scan(&playerCount, &maxPlayers)
	switch {
	case err == sql.ErrNoRows:
		return nil, ErrGameNotExist
	case err != nil:
		log.Print(err)
		return nil, err
	}

	if playerCount >= maxPlayers {
		return nil, ErrGameFull
	}

	var character model.Character
	character.CharacterId = characterId

	var gameData string

	err = db.QueryRow("SELECT name, last_game_id, game_data FROM characters WHERE id = $1 AND uid = $2", characterId, userId).Scan(&character.Name, &character.LastGameId, &gameData)
	if err != nil {
		return nil, err
	}

	var state model.CharacterState
	err = json.Unmarshal([]byte(gameData), &state)
	if err != nil {
		return nil, err
	}

	character.CharacterState = state

	// increment playercount
	_, err = db.Exec("UPDATE games SET player_count = $1 WHERE game_id = $2", playerCount+1, gameId)
	if err != nil {

		return nil, err
	}

	return &character, nil
}

func GetCharacter(machineKey string, characterId int) (*model.Character, error) {

	machineId, err := readMachineKey(machineKey)
	if err != nil {

		return nil, err
	}

	var realMachineKey string
	err = db.QueryRow("SELECT machine_key FROM machines WHERE machine_id = $1", machineId).Scan(&realMachineKey)
	if err != nil {
		return nil, err
	}

	if machineKey != realMachineKey {

		return nil, ErrInvalidMachineKey
	}

	var character model.Character
	character.CharacterId = characterId

	var gameData string

	err = db.QueryRow("SELECT name, last_game_id, game_data FROM characters WHERE id = $1", characterId).Scan(&character.Name, &character.LastGameId, &gameData)
	if err != nil {
		return nil, err
	}

	var state model.CharacterState
	err = json.Unmarshal([]byte(gameData), &state)
	if err != nil {
		return nil, err
	}

	character.CharacterState = state

	return &character, nil
}

func UpdateCharacter(machineKey string, character *model.Character) error {

	machineId, err := readMachineKey(machineKey)
	if err != nil {

		return err
	}

	var realMachineKey string
	err = db.QueryRow("SELECT machine_key FROM machines WHERE machine_id = $1", machineId).Scan(&realMachineKey)
	if err != nil {
		return err
	}

	if machineKey != realMachineKey {

		return errors.New("invalid machine key")
	}

	json, err := json.Marshal(&character.CharacterState)
	if err != nil {

		return err
	}

	_, err = db.Exec("UPDATE characters SET last_game_id = $1, game_data = $2 WHERE id = $3", character.LastGameId, string(json), character.CharacterId)
	if err != nil {

		return err
	}

	return nil
}

func GetGamesList() ([]model.Game, error) {

	rows, err := db.Query("SELECT * FROM games")
	if err != nil {
		return nil, err
	}

	list := make([]model.Game, 0)

	for rows.Next() {
		var game model.Game
		err = rows.Scan(&game.GameId, &game.Map, &game.Mode, &game.MinimumLevel, &game.PlayerCount, &game.MaximumPlayers)
		if err != nil {
			log.Print("game read error:", err)
		} else {
			list = append(list, game)
		}
	}

	return list, nil
}

func GetMachineList() ([]model.Machine, error) {

	rows, err := db.Query("SELECT machine_id, remote_address, service_listen_port, most_recent_key FROM machines JOIN machines_metadata USING (machine_id)")
	if err != nil {
		return nil, err
	}

	list := make([]model.Machine, 0)

	for rows.Next() {
		var m model.Machine
		err = rows.Scan(&m.MachineId, &m.RemoteAddress, &m.ListenPort, &m.MachineKey)
		if err != nil {
			log.Print("machine read error:", err)
		} else {
			list = append(list, m)
		}
	}

	return list, nil

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

func validateMachineKey(machineKey string) (machineId int, valid bool, err error) {

	machineId, err = readMachineKey(machineKey)
	if err != nil {

		return
	}

	var realMachineKey string
	err = db.QueryRow("SELECT most_recent_key FROM machines JOIN machines_metadata USING (machine_id) WHERE machine_id = $1", machineId).Scan(&realMachineKey)
	if err != nil {

		return
	}

	if machineKey != realMachineKey {

		return 0, false, ErrInvalidMachineKey
	}

	return machineId, true, nil
}
