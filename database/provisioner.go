package thordb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"github.com/jaybennett89/thorium-go/requests"
)

func ProvisionNewGame(game_id int, map_name string, game_mode string) error {

	log.Print("starting new game on %s (%s)", map_name, game_mode)

	var (
		address      string
		port         int
		machineToken string
	)

	err := db.QueryRow("SELECT * FROM get_available_machine()").Scan(&address, &port, &machineToken)
	if err != nil {
		log.Print("no available machines")
		return errors.New("thordb: does not exist")
	}

	var data request.NewGameServer
	//	data.MachineToken = machineToken
	//	data.MapName = map_name
	//	data.GameMode = game_mode

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(&data)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("http://%s:%d/games/%d", address, port, game_id)
	log.Printf("provisioner: posting new game to @%s", endpoint)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	var response *http.Response
	response, err = client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("thordb: failed to initialize resource")
	} else {
		log.Print("provisioner: new game request ok")
	}
	return nil
}

func ProvisionNewMachine() {
	// todo: spawn a new machine in aws or similar
}
