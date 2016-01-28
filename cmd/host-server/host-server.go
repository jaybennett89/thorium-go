package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"thorium-go/client"
	"thorium-go/launch"
	"thorium-go/requests"
	"thorium-go/usage"
	"time"
)

import _ "github.com/lib/pq"
import "github.com/go-martini/martini"

// application data
var registerData request.MachineRegisterResponse
var listenPort int

var masterEndpoint string = "thorium-sky.net:6960"

func main() {
	fmt.Println("hello world")

	timeNow := time.Now()
	rand.Seed(int64(timeNow.Second()))
	listenPort = rand.Intn(50000)
	listenPort = listenPort + 10000

	fmt.Println(strconv.Itoa(listenPort), "\n")

	reqData := &request.RegisterMachine{Port: listenPort}
	jsonBytes, err := json.Marshal(reqData)
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprintf("http://%s/machines/register", masterEndpoint)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if response.StatusCode != 200 {
		log.Print("Error registering with master")
		os.Exit(1)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	json.Unmarshal([]byte(body), &registerData)
	fmt.Println("Registered As Machine#", registerData.MachineId)

	m := martini.Classic()

	// called by master
	m.Get("/", handlePingRequest)
	m.Get("/status", handlePingRequest)
	m.Post("/games", handlePostNewGame)

	// called by local gameservers
	m.Post("/games/register_server", handleRegisterLocalServer)
	m.Post("/games/player_connect", handlePlayerConnect)
	m.Post("/games/player_disconnect", handlePlayerDisconnect)
	m.Post("/characters", handleUpdateCharacter)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		fmt.Println(<-c)
		shutdown()
		os.Exit(1)
	}()
	defer shutdown()

	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHeartbeat()
			}
		}
	}()

	thisIp := fmt.Sprintf(":%d", listenPort)
	m.RunOnAddr(thisIp)
}

func sendHeartbeat() {
	var err error
	statusData := &request.MachineStatus{}
	statusData.MachineKey = registerData.MachineKey
	statusData.UsageCPU, _ = usage.GetCPU()
	statusData.UsageNetwork, _ = usage.GetNetworkUtilization()
	statusData.PlayerCapacity = 0.0

	jsonBytes, err := json.Marshal(statusData)
	if err != nil {

		log.Print(err)
		return
	}

	url := fmt.Sprintf("http://%s/machines/status", masterEndpoint)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {

		log.Print(err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Connection", "close")

	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(request)
	if err != nil {

		log.Print(err)
		return
	}

	resp.Body.Close()
}

func handlePingRequest() (int, string) {
	return 200, "OK"
}

func handlePostNewGame(httpReq *http.Request, params martini.Params) (int, string) {

	defer httpReq.Body.Close()

	var data request.NewGameServer
	decoder := json.NewDecoder(httpReq.Body)

	err := decoder.Decode(&data)
	if err != nil {

		log.Print("bad json request:\n", httpReq.Body)
		return 400, err.Error() // okay to send err back to master
	}

	err = launch.NewGameServer(registerData.MachineKey, listenPort, data.GameId, data.Map, data.Mode, data.MinimumLevel, data.MaximumPlayers)
	if err != nil {

		log.Print(err)
		return 500, "Internal Server Error"
	}

	response := request.NewGameServerResponse{registerData.MachineKey}

	json, err := json.Marshal(&response)
	if err != nil {

		return 500, err.Error()
	}

	return 200, string(json)
}

func handlePlayerConnect(httpReq *http.Request) (int, string) {

	var data request.PlayerConnect
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&data)
	if err != nil {

		fmt.Println(err)
		return 400, "Bad Request"
	}

	if data.MachineKey != registerData.MachineKey {

		log.Print("WARNING: Received invalid machine key during player connect")
		log.Printf("have %s recv %s", registerData.MachineKey, data.MachineKey)
		return 403, "Invalid Key"
	}

	rc, body, err := client.PlayerConnect(masterEndpoint, data.GameId, data.MachineKey, data.SessionKey, data.CharacterId)

	if err != nil {

		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	return rc, body
}

func handlePlayerDisconnect(httpReq *http.Request) (int, string) {

	var data request.PlayerDisconnect
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&data)
	if err != nil {

		fmt.Println(err)
		return 400, "Bad Request"
	}

	if data.MachineKey != registerData.MachineKey {

		log.Print("WARNING: Received invalid machine key during player connect")
		log.Printf("have %s recv %s", registerData.MachineKey, data.MachineKey)
		return 403, "Invalid Key"
	}

	rc, body, err := client.PlayerDisconnect(masterEndpoint, data.MachineKey, data.GameId, data.Snapshot)

	if err != nil {

		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	return rc, body
}

func handleUpdateCharacter(httpReq *http.Request) (int, string) {

	var data request.UpdateCharacter
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&data)
	if err != nil {

		fmt.Println(err)
		return 400, "Bad Request"
	}

	if data.MachineKey != registerData.MachineKey {

		log.Print("WARNING: Received invalid machine key during update character")
		log.Printf("have %s recv %s", registerData.MachineKey, data.MachineKey)
		return 403, "Invalid Key"
	}

	rc, body, err := client.UpdateCharacter(masterEndpoint, data.MachineKey, data.Snapshot)

	if err != nil {

		fmt.Println(err)
		return 500, "Internal Server Error"
	}

	return rc, body
}

func handleRegisterLocalServer(httpReq *http.Request, params martini.Params) (int, string) {

	decoder := json.NewDecoder(httpReq.Body)
	var data request.RegisterGameServer
	err := decoder.Decode(&data)
	if err != nil {

		log.Print(err)
		return 400, "Bad Request"
	}

	if data.MachineKey != registerData.MachineKey {

		log.Print("WARNING: Received invalid key trying to register local gameserver")
		log.Printf("have %s recv %s", registerData.MachineKey, data.MachineKey)
		return 403, "Invalid Key"
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&data)
	if err != nil {

		log.Print(err)
		return 500, "Internal Server Error"
	}

	endpoint := fmt.Sprintf("http://%s/games/register_server", masterEndpoint)
	var req *http.Request
	req, err = http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {

		log.Print(err)
		return 500, "Internal Server Error"
	}

	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {

		log.Print(err)
		return 500, "Internal Server Error"
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {

		log.Print("error: couldn't register game server with master")
		return 400, "Bad Request"
	}

	return 200, "OK"
}

func shutdown() {

	var reqData request.UnregisterMachine
	reqData.MachineKey = registerData.MachineKey
	jsonBytes, err := json.Marshal(&reqData)
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest("POST", fmt.Sprintf("http://52.25.124.72:6960/machines/%d/disconnect", registerData.MachineId), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		log.Print("failed to disconnect properly")
		return
	}
	resp.Body.Close()
}
