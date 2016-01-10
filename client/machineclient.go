package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"thorium-go/requests"
)

func NewGameServer(endpoint string, gameId int, mapName string, mode string, minLevel int, maxPlayers int) (int, string, error) {

	data := request.NewGameServer{
		GameId:         gameId,
		Map:            mapName,
		Mode:           mode,
		MinimumLevel:   minLevel,
		MaximumPlayers: maxPlayers,
	}

	jsonBytes, err := json.Marshal(&data)
	if err != nil {
		return 0, "", err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/games", endpoint), bytes.NewBuffer(jsonBytes))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Error with request: ", err)
		return 0, "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}
