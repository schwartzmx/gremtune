package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/schwartzmx/gremtune"
)

var panicOnErrorOnChannel = func(errs chan error) {
	err := <-errs
	if err == nil {
		return // ignore if the channel was closed
	}
	log.Fatalf("Lost connection to the database: %s", err)
}

func main() {

	host := "localhost"
	port := 8182
	hostURL := fmt.Sprintf("ws://%s:%d", host, port)

	errs := make(chan error)
	go panicOnErrorOnChannel(errs)

	dialer := gremtune.NewWebsocketDialer(hostURL)    // Returns a WebSocket dialer to connect to Gremlin Server
	gremlinClient, err := gremtune.Dial(dialer, errs) // Returns a gremtune client to interact with
	if err != nil {
		log.Fatalf("Failed to create the gremlin client: %s", err)
	}

	// Sends a query to Gremlin Server
	res, err := gremlinClient.Execute("g.V()")
	if err != nil {
		log.Fatalf("Failed to execute a gremlin command: %s", err)
	}

	jsonEncodedResponse, err := json.MarshalIndent(res[0].Result.Data, "", "    ")
	if err != nil {
		log.Fatalf("Failed to encode the raw json into json: %s", err)
	}

	log.Printf("Received data: \n%s\n", jsonEncodedResponse)
}
