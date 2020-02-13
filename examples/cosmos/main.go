package main

import (
	"encoding/json"
	"log"
	"os"

	gremcos "github.com/supplyon/gremcos"
)

var panicOnErrorOnChannel = func(errs chan error) {
	err := <-errs
	if err == nil {
		return // ignore if the channel was closed
	}
	log.Fatalf("Lost connection to the database: %s", err)
}

func main() {

	host := os.Getenv("CDB_HOST")
	username := os.Getenv("CDB_USERNAME")
	password := os.Getenv("CDB_KEY")

	if len(host) == 0 {
		log.Fatal("Host not set. Use export CDB_HOST=<CosmosDB Gremlin Endpoint> to specify it")
	}

	if len(username) == 0 {
		log.Fatal("Username not set. Use export CDB_USERNAME=/dbs/<cosmosdb name>/colls/<graph name> to specify it")
	}

	if len(password) == 0 {
		log.Fatal("Key not set. Use export CDB_KEY=<key> to specify it")
	}

	log.Println("Connecting using:")
	log.Printf("\thost: %s\n", host)
	log.Printf("\tusername: %s\n", username)
	log.Printf("\tpassword is set %v\n", len(password) > 0)

	errs := make(chan error)
	go panicOnErrorOnChannel(errs)

	websocket, err := gremcos.NewWebsocket(host)
	if err != nil {
		log.Fatalf("Failed to create the websocket: %s", err)
	}

	gremlinClient, err := gremcos.Dial(websocket, errs, gremcos.SetAuth(username, password)) // Returns a gremcos client to interact with
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
