package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"thorium-go/requests"
)

import "bytes"
import "io/ioutil"

var address string = "52.25.124.72"
var port int = 6960

func GetStatus(masterEndpoint string) (int, string, error) {

	url := fmt.Sprintf("http://%s/status", masterEndpoint)
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		return 0, "", err
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Print("ping master - error:\n", err)
		return 0, "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(bodyBytes), nil
}

func Register(masterEndpoint string, username string, password string) (int, string, error) {

	// create request data struct in memory
	var loginReq request.Authentication
	loginReq.Username = username
	loginReq.Password = password

	// marshal request data into json byte array
	jsonBytes, err := json.Marshal(&loginReq)
	if err != nil {
		return 0, "", err
	}

	// create the http request struct
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/clients/register", address, port), bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Print("error with request: ", err)
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// create the http client struct and execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("error with sending request", err)
		return 0, "", err
	}

	// read the body into a byte array and return the results
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func Login(masterEndpoint string, username string, password string) (int, string, error) {

	// create request data struct in memory
	var loginReq request.Authentication
	loginReq.Username = username
	loginReq.Password = password

	// marshal request data into json byte array
	jsonBytes, err := json.Marshal(&loginReq)
	if err != nil {
		return 0, "", err
	}

	// create the http request struct
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/clients/login", address, port), bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Print("error with request: ", err)
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// create the http client struct and execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("error with sending request", err)
		return 0, "", err
	}

	// read the body into a byte array and return the results
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}

	return resp.StatusCode, string(body), nil
}

func CharacterSelectRequest(token string, id int) (string, error) {

	var selectReq request.SelectCharacter
	selectReq.AccountToken = token
	selectReq.ID = id
	jsonBytes, err := json.Marshal(&selectReq)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/characters/%d/select", address, port, id), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Error with request 2: ", err)
		return "err", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func CharacterCreateRequest(token string, name string) (string, error) {

	var charCreateReq request.CreateCharacter
	charCreateReq.AccountToken = token
	charCreateReq.Name = name
	jsonBytes, err := json.Marshal(&charCreateReq)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/characters/new", address, port), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Error with request 2: ", err)
		return "err", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print("create character response: ", string(body))
	return string(body), nil
}

func DisconnectRequest(token string) (string, error) {

	buf := []byte(token)
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/clients/disconnect", address, port), bytes.NewBuffer(buf))
	if err != nil {
		log.Print("error with request: ", err)
		return "err", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("error with sending request", err)
		return "err", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Print("disonnect response: ", string(body))
	return string(body), nil
}
