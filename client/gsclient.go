package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"thorium-go/model"
	"thorium-go/requests"
)

// this package contains game server requests

func PlayerConnect(serviceEndpoint string, gameId int, machineKey string, sessionKey string, characterId int) (statusCode int, body string, err error) {

	data := request.PlayerConnect{
		GameId:      gameId,
		MachineKey:  machineKey,
		SessionKey:  sessionKey,
		CharacterId: characterId}

	json, err := json.Marshal(&data)
	if err != nil {

		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/games/player_connect", serviceEndpoint), bytes.NewBuffer(json))
	if err != nil {

		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes), nil
}

func UpdateCharacter(serviceEndpoint string, machineKey string, character *model.Character) (statusCode int, body string, err error) {

	data := request.UpdateCharacter{
		MachineKey: machineKey,
		Snapshot:   character}

	json, err := json.Marshal(&data)
	if err != nil {

		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/characters", serviceEndpoint), bytes.NewBuffer(json))
	if err != nil {

		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes), nil
}

func PlayerDisconnect(serviceEndpoint string, machineKey string, gameId int, character *model.Character) (statusCode int, body string, err error) {

	data := request.PlayerDisconnect{
		MachineKey: machineKey,
		GameId:     gameId,
		Snapshot:   character}

	json, err := json.Marshal(&data)
	if err != nil {

		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/games/player_disconnect", serviceEndpoint), bytes.NewBuffer(json))
	if err != nil {

		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes), nil
}
