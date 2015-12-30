package client

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"thorium-go/requests"
	"time"
)

var masterEndpoint = "thorium-sky.net:6960"
var accountToken string

var user string = "test"
var password string = "test"

// Test 1A: Service Status
// HTTP GET /status

func Test1A_Ping(t *testing.T) {
	statusCode, _, err := GetStatus(masterEndpoint)

	if err != nil {
		t.Fail()
	} else if statusCode != 200 {
		log.Printf("recieved bad status code %d", statusCode)
		t.Fail()
	} else {
		log.Print("Test 1A Success")
	}
}

// Test 2A: Register
// HTTP POST /clients/register

func Test2A_Register(t *testing.T) {

	fmt.Println("Test 2A: Register")

	tries := 3

	for tries > 0 {

		// pick a random number between 10k-1M
		rand.Seed(time.Now().UTC().UnixNano())
		randNum := rand.Intn(990000) + 10000
		// create user string
		user = fmt.Sprintf("test%d", randNum)

		fmt.Println(user)
		// execute request
		responseCode, body, err := Register(masterEndpoint, user, password)
		if err != nil {
			log.Print(err)
			t.FailNow()
		}

		fmt.Printf("register response: status %d\n", responseCode)
		if responseCode == 200 {
			log.Print("registered ", user)

			var resp request.LoginResponse
			json.Unmarshal([]byte(body), &resp)

			log.Print("userToken=", resp.UserToken)
			if resp.UserToken != "" {
				// success
				accountToken = resp.UserToken
				return
			}
		}

		tries = tries - 1
	}

	// fail if all tries are used up
	log.Print("all register tries used")
	t.Fail()
}

// Test 2B: Disconnect
// HTTP POST /clients/disconnect
func Test2B_Disconnect(t *testing.T) {

	fmt.Println("Test 2B: Disconnect")

	rc, _, err := Disconnect(masterEndpoint, accountToken)
	if err != nil {
		log.Print(err)
		t.FailNow()
	}

	log.Print("disconnect response code = ", rc)
	if rc != 200 {
		t.FailNow()
	}
}

// Test 2B: Login
// HTTP POST /clients/login

func Test2B_Login(t *testing.T) {
	fmt.Println("Test 2B: Login")

	// execute request
	responseCode, body, err := Login(masterEndpoint, user, password)
	if err != nil {
		log.Print(err)
		t.FailNow()
	}

	fmt.Printf("register response: status %d, %s\n", responseCode, body)
	if responseCode != 200 {
		t.Fail()
	} else {
		log.Print("login ", user)
		log.Print("response body ", body)
	}
}
