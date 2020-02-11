package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune"
)

func main() {

	host := "localhost"
	port := 8182
	hostURL := fmt.Sprintf("ws://%s:%d/gremlin", host, port)
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{})

	cosmos, err := gremtune.New(hostURL, logger)
	if err != nil {
		log.Fatalf("Failed to create the cosmos connector: %s", err)
	}

	for {

		res, err := cosmos.Execute("g.addV('Phil')")
		if err != nil {
			log.Fatalf("Failed to execute a gremlin command: %s", err)
		}

		jsonEncodedResponse, err := json.MarshalIndent(res[0].Result.Data, "", "    ")
		if err != nil {
			log.Fatalf("Failed to encode the raw json into json: %s", err)
		}

		log.Printf("Received data: \n%s\n", jsonEncodedResponse)
		time.Sleep(time.Second * 1)
	}
}
