package client

import (
	"log"
	"testing"
)

var masterEndpoint = "thorium-sky.net:6960"

// Test 1A: Service Status Response
// HTTP GET /status

func Test1A_Ping(t *testing.T) {
	statusCode, _, err := GetStatus(masterEndpoint)

	if err != nil {
		t.Fail()
	} else if statusCode != 200 {
		log.Printf("recieved bad status code %d", statusCode)
		t.Fail()
	}
}

// Test 2A:
