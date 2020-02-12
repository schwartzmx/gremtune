package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/schwartzmx/gremtune"
	"github.com/schwartzmx/gremtune/interfaces"
)

func main() {

	host := "localhost"
	port := 8182
	hostURL := fmt.Sprintf("ws://%s:%d/gremlin", host, port)
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}).With().Timestamp().Logger()

	cosmos, err := gremtune.New(hostURL, gremtune.WithLogger(logger), gremtune.NumMaxActiveConnections(10), gremtune.ConnectionIdleTimeout(time.Second*1))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create the cosmos connector")
	}

	exitChannel := make(chan struct{})
	go processLoop(cosmos, logger, exitChannel)

	<-exitChannel
	if err := cosmos.Stop(); err != nil {
		logger.Error().Err(err).Msg("Failed to stop")
	}
	logger.Info().Msg("Teared down")
}

func processLoop(cosmos *gremtune.Cosmos, logger zerolog.Logger, exitChannel chan<- struct{}) {
	// register for common exit signals (e.g. ctrl-c)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// create tickers for doing health check and queries
	queryTicker := time.NewTicker(time.Second * 2)
	healthCheckTicker := time.NewTicker(time.Second * 1)

	// ensure to clean up as soon as the processLoop has been left
	defer func() {
		queryTicker.Stop()
		healthCheckTicker.Stop()
	}()

	stopProcessing := false
	logger.Info().Msg("Process loop entered")
	for !stopProcessing {
		select {
		case <-signalChannel:
			close(exitChannel)
			stopProcessing = true
		case <-queryTicker.C:
			queryCosmos(cosmos, logger)
		case <-healthCheckTicker.C:
			logger.Debug().Bool("healthy", cosmos.IsHealthy()).Msg("Health Check")
		}
	}

	logger.Info().Msg("Process loop left")
}

func queryCosmos(cosmos *gremtune.Cosmos, logger zerolog.Logger) {
	res, err := cosmos.Execute("g.addV('Phil')")
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute a gremlin command")
		return
	}

	for i, chunk := range res {
		jsonEncodedResponse, err := json.Marshal(chunk.Result.Data)

		if err != nil {
			logger.Error().Err(err).Msg("Failed to encode the raw json into json")
			continue
		}
		logger.Info().Str("reqID", chunk.RequestID).Int("chunk", i).Msgf("Received data: %s", jsonEncodedResponse)
	}
}

func queryCosmosAsync(cosmos *gremtune.Cosmos, logger zerolog.Logger) {
	dataChannel := make(chan interfaces.AsyncResponse)

	go func() {
		for chunk := range dataChannel {
			jsonEncodedResponse, err := json.Marshal(chunk.Response.Result.Data)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to encode the raw json into json")
				continue
			}
			logger.Info().Str("reqID", chunk.Response.RequestID).Msgf("Received data: %s", jsonEncodedResponse)
			time.Sleep(time.Millisecond * 200)
		}
	}()

	err := cosmos.ExecuteAsync("g.addV('Phil')", dataChannel)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute async a gremlin command")
		return
	}
}
