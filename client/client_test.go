package client

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"
)

var masterEndpoint = "thorium-sky.net:6960"
var accountToken = ""

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
		fmt.Println(randNum)
		// create user string
		user = fmt.Sprintf("test%d", randNum)

		fmt.Println(user)
		// execute request
		responseCode, accountToken, err := Register(masterEndpoint, user, password)
		if err != nil {
			log.Print(err)
			t.FailNow()
		}

		fmt.Printf("register response: status %d, %s\n", responseCode, accountToken)
		if responseCode == 200 {
			log.Print("registered ", user)
			log.Print("response body ", accountToken)
			return
		}

		tries = tries - 1
	}

	t.Fail()
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

	fmt.Printf("register response: status %d, %s\n", responseCode, accountToken)
	if responseCode == 200 {
		log.Print("login ", user)
		log.Print("response body ", body)
		return
	}

	//statusCode, body, err := Login(masterEndpoint, user, password)
}
