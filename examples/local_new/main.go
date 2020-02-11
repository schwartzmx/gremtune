package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune"
)

func main() {

	host := "localhost"
	port := 8182
	hostURL := fmt.Sprintf("ws://%s:%d/gremlin", host, port)
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}).With().Timestamp().Logger()

	cosmos, err := gremtune.New(hostURL, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create the cosmos connector")
	}

	for {

		time.Sleep(time.Second * 1)

		res, err := cosmos.Execute("g.addV('Phil')")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to execute a gremlin command")
			continue
		}

		jsonEncodedResponse, err := json.MarshalIndent(res[0].Result.Data, "", "    ")
		if err != nil {
			logger.Error().Err(err).Msg("Failed to encode the raw json into json")
			continue
		}

		logger.Info().Msgf("Received data: \n%s\n", jsonEncodedResponse)
	}
}
