package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"thorium-go/model"
	"thorium-go/requests"
	"time"

	"github.com/go-martini/martini"
)

func main() {
	log.Print("running a mock game server")

	var key string
	var listenPort int
	var servicePort int
	var game model.Game

	flag.StringVar(&key, "key", "", "machine key")
	flag.IntVar(&game.GameId, "id", 0, "identifies this game within the cluster")
	flag.IntVar(&listenPort, "listen", 0, "game server listen port")
	flag.IntVar(&servicePort, "service", 0, "machine local service port")
	flag.StringVar(&game.Map, "map", "mp_sandbox", "game map: mp_sandbox, mp_openworld")
	flag.StringVar(&game.Mode, "mode", "tutorial", "game mode: basic, tutorial, freeforall")
	flag.IntVar(&game.MinimumLevel, "minlvl", 0, "minimum level of player")
	flag.IntVar(&game.MaximumPlayers, "maxplayers", 16, "maximum player count")

	flag.Parse()

	if game.GameId == 0 || listenPort == 0 || servicePort == 0 || game.Map == "" || game.Mode == "" {
		log.Fatal("bad arguments")
	}

	var data request.RegisterGameServer
	data.MachineKey = key
	data.Port = listenPort
	data.GameId = game.GameId
	jsonBytes, err := json.Marshal(&data)

	time.Sleep(180 * time.Millisecond)

	endpoint := fmt.Sprintf("http://localhost:%d/games/register_server", servicePort)
	var req *http.Request
	req, err = http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("Die - failed to register")
	}

	// todo: service heartbeat and client hearbeat listener

	m := martini.Classic()
	m.Get("/status", handleStatusRequest)
	m.Post("/connect", handleConnectRequest)
	m.RunOnAddr(fmt.Sprintf(":%d", listenPort))
}

func handleStatusRequest(httpReq *http.Request) (int, string) {
	return 200, "OK"
}

func handleConnectRequest(httpReq *http.Request) (int, string) {

	return 200, "OK"
}
