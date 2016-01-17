package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)
import "github.com/go-martini/martini"
import (
	"thorium-go/database"
	"thorium-go/requests"
)

func main() {
	fmt.Println("hello world")

	m := martini.Classic()

	// status
	m.Get("/", handleGetStatusRequest)
	m.Get("/status", handleGetStatusRequest)

	// client
	m.Post("/clients/login", handleClientLogin)
	m.Post("/clients/register", handleClientRegister)
	m.Post("/clients/disconnect", handleClientDisconnect)

	// characters
	m.Post("/characters/new", handleCreateCharacter)
	m.Get("/characters/select", handleSelectCharacter)
	m.Get("/characters/:id/profile", handleGetCharProfile)

	// games
	m.Post("/games/register_server", handleRegisterServer)
	m.Post("/games/server_status", handleGameServerStatus)

	m.Post("/games", handleNewGameRequest)

	m.Get("/games", handleGetServerList)
	m.Get("/games/:id", handleGetGameInfo)
	m.Get("/games/:id/server_info", handleGetServerInfo)
	m.Post("/games/join", handleClientJoinGame)
	m.Post("/games/join_queue", handleClientJoinQueue)

	// machines
	m.Post("/machines/register", handleRegisterMachine)
	m.Post("/machines/status", handleMachineHeartbeat)
	m.Post("/machines/:id/disconnect", handleUnregisterMachine)
	m.Delete("/machines/:id", handleUnregisterMachine)

	m.RunOnAddr(":6960")
}

func handleGetStatusRequest(httpReq *http.Request) (int, string) {
	return 200, "OK"
}

func handleClientLogin(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.Authentication
	err := decoder.Decode(&req)
	if err != nil {
		log.Print("bad json request", httpReq.Body)
		return 400, "Bad Request"
	}

	var username string
	var password string
	username, password, err = sanitize(req.Username, req.Password)
	if err != nil {
		log.Print("Error sanitizing authentication request", req.Username, req.Password)
		return 400, "Bad Request"
	}

	var charIDs []int
	var token string
	token, charIDs, err = thordb.LoginAccount(username, password)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: does not exist":
			log.Print(fmt.Sprintf("thordb: failed login attempt: %s//%s", username, password))
			return 400, "Bad Request"
		case "thordb: invalid password":
			log.Print(fmt.Sprintf("thordb: failed login attempt: %s//%s", username, password))
			return 400, "Bad Request"
		case "thordb: already logged in":
			log.Printf("thordb: failed login attempt (already logged in): %s//%s", username, password)
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}

	var resp request.LoginResponse
	resp.SessionKey = token
	resp.CharacterIDs = charIDs
	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&resp)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}
	return 200, string(jsonBytes)
}

func handleClientRegister(httpReq *http.Request) (int, string) {
	//using authentication struct for now because i haven't added the token yet
	var req request.Authentication
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("error decoding register account request (authentication)")
		return 500, "Internal Server Error"
	}

	var username string
	var password string

	username, password, err = sanitize(req.Username, req.Password)
	if err != nil {
		log.Print("Error sanitizing authentication request", req.Username, req.Password)
		return 400, "Bad Request"
	}

	token, charIds, err := thordb.RegisterAccount(username, password)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: already in use":
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}

	var resp request.LoginResponse
	resp.SessionKey = token
	resp.CharacterIDs = charIds
	jsonBytes, err := json.Marshal(&resp)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	return 200, string(jsonBytes)
}

func handleClientDisconnect(httpReq *http.Request) (int, string) {

	// getting session key might become more complicated later (add a request struct)
	bytes, err := ioutil.ReadAll(httpReq.Body)
	if err != nil {
		return 400, "Bad Request"
	}

	accountSession := string(bytes)
	err = thordb.Disconnect(accountSession)
	if err != nil {
		log.Print("thordb couldnt disconnect, something went wrong")
		log.Print(err)
		return 400, "Bad Request"
	}

	return 200, "OK"
}

func handleCreateCharacter(httpReq *http.Request) (int, string) {
	var req request.CreateCharacter
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Print("character create req json decoding error %s", httpReq.Body)
		return 400, "Bad Request"
	}

	characterId, err := thordb.CreateCharacter(req.SessionKey, req.Name, req.ClassId)
	if err != nil {
		log.Print(err)
		switch err.Error() {
		case "thordb: already in use":
			return 400, "Bad Request"
		case "token contains an invalid number of segments":
			return 400, "Bad Request"
		default:
			return 500, "Internal Server Error"
		}
	}

	var resp request.NewCharacterResponse
	resp.CharacterId = characterId

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&resp)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	return 200, string(jsonBytes)
}

func handleSelectCharacter(httpReq *http.Request) (int, string) {

	var req request.SelectCharacter
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Print("character select req json decoding error %s", httpReq.Body)
		return 400, "Bad Request"
	}

	character, err := thordb.SelectCharacter(req.SessionKey, req.CharacterId)
	if err != nil {
		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	json, err := json.Marshal(&character)
	if err != nil {
		return 500, "Internal Server Error"
	}

	return 200, string(json)
}

func handleGetCharProfile(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleClientJoinGame(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleClientJoinQueue(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGameServerStatus(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleGetServerList(httpReq *http.Request) (int, string) {

	list, err := thordb.GetGamesList()
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	bytes, err := json.Marshal(list)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	return 200, string(bytes)
}

func handleGetGameInfo(httpReq *http.Request) (int, string) {
	return 500, "Not Implemented"
}

func handleRegisterMachine(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.RegisterMachine
	err := decoder.Decode(&req)
	if err != nil {
		logerr("Error decoding machine register request", err)
		return 500, "Internal Server Error"
	}

	if req.Port == 0 {
		fmt.Println("No Port Given")
		return 400, "Bad Request"
	} else {
		fmt.Println("register port = ", req.Port)
	}

	machineIp := strings.Split(httpReq.RemoteAddr, ":")[0]

	var machineId int
	var machineKey string
	machineId, machineKey, err = thordb.RegisterMachine(machineIp, req.Port)
	if err != nil {
		logerr("error registering machine", err)
		return 500, "Internal Server Error"
	}
	var response request.MachineRegisterResponse
	response.MachineId = machineId
	response.MachineKey = machineKey

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&response)
	if err != nil {
		log.Print("error encoding register machine response\n", err)
		return 500, "Internal Server Error"
	}

	return 200, string(jsonBytes)
}

func handleUnregisterMachine(httpReq *http.Request, params martini.Params) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.UnregisterMachine
	err := decoder.Decode(&req)
	if err != nil {
		logerr("Error decoding machine unregister request", err)
		return 400, "Bad Request"
	}

	success, err := thordb.UnregisterMachine(req.MachineKey)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	} else if !success {
		logerr("unable to remove machine registry", err)
		return 400, "Bad Request"
	}

	return 200, "OK"
}

func handleNewGameRequest(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.CreateNewGame
	err := decoder.Decode(&req)
	if err != nil {
		logerr("unable to decode body data", err)
		return 500, "Internal Server Error"
	}

	if req.Map == "" || req.GameMode == "" {
		return 400, "Missing Parameters"
	}

	if req.MaxPlayers == 0 || req.MaxPlayers > 64 {
		req.MaxPlayers = 16
		// use default in extreme cases
	}

	// validate token

	var gameId int
	gameId, err = thordb.CreateNewGame(req.Map, req.GameMode, req.MinimumLevel, req.MaxPlayers)
	if err != nil {

		fmt.Println(err)

		switch err.Error() {

		case "thordb: no available servers":
			return 503, "No Available Servers"
		default:
			return 500, "Internal Server Error"
		}

	}

	response := request.CreateNewGameResponse{GameId: gameId}
	bytes, err := json.Marshal(&response)
	if err != nil {

		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	fmt.Println("[ThoriumNET] new game, id=", strconv.Itoa(gameId))
	return 201, string(bytes)
}

func handleRegisterServer(httpReq *http.Request, params martini.Params) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.RegisterGameServer
	err := decoder.Decode(&req)
	if err != nil {

		logerr("Error decoding machine register request", err)
		return 500, "Internal Server Error"
	}

	if req.Port == 0 {

		fmt.Println("No Port Given")
		return 400, "Missing Parameters"
	}

	err = thordb.RegisterActiveGame(req.GameId, req.MachineKey, req.Port)
	if err != nil {

		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	return 200, "OK"
}

func handleMachineHeartbeat(httpReq *http.Request) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var req request.MachineStatus
	err := decoder.Decode(&req)
	if err != nil || req.MachineKey == "" {
		log.Print("bad json request", httpReq.Body)
		return 400, "Bad Request"
	}

	err = thordb.UpdateMachineStatus(req.MachineKey, req.UsageCPU, req.UsageNetwork, req.PlayerCapacity)
	if err != nil {
		log.Print(err)
		return 500, "Internal Server Error"
	}

	return 200, "OK"
}

func handleGetServerInfo(params martini.Params) (int, string) {

	gameId, err := strconv.Atoi(params["id"])
	if err != nil {
		return 400, "Bad Request"
	}

	host, running, err := thordb.GetServerInfo(gameId)
	switch {

	case err == thordb.GameNotExistError:

		return 404, "Game Not Found"

	case err != nil:

		fmt.Println(err)
		return 500, "Internal Server Error"

	}

	if !running {

		return 202, "Accepted"
	}

	var data request.ServerInfoResponse
	data.RemoteAddress = host.RemoteAddress
	data.ListenPort = host.ListenPort

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&data)
	if err != nil {
		return 500, "Internal Server Error"
	}

	return 200, string(jsonBytes)
}

func sanitize(username string, password string) (string, string, error) {
	return username, password, nil
}

// TODO: Refactor into logging package
func logerr(msg string, err error) {
	fmt.Println("[ThoriumNET] ", msg)
	fmt.Println(err)
}
