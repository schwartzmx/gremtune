package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	gremcos "github.com/supplyon/gremcos"
	"github.com/supplyon/gremcos/api"
)

func main() {

	host := os.Getenv("CDB_HOST")
	username := os.Getenv("CDB_USERNAME")
	password := os.Getenv("CDB_KEY")
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}).With().Timestamp().Logger()

	if len(host) == 0 {
		logger.Fatal().Msg("Host not set. Use export CDB_HOST=<CosmosDB Gremlin Endpoint> to specify it")
	}

	if len(username) == 0 {
		logger.Fatal().Msg("Username not set. Use export CDB_USERNAME=/dbs/<cosmosdb name>/colls/<graph name> to specify it")
	}

	if len(password) == 0 {
		logger.Fatal().Msg("Key not set. Use export CDB_KEY=<key> to specify it")
	}

	log.Println("Connecting using:")
	log.Printf("\thost: %s\n", host)
	log.Printf("\tusername: %s\n", username)
	log.Printf("\tpassword is set %v\n", len(password) > 0)

	cosmos, err := gremcos.New(host,
		gremcos.WithAuth(username, password),
		gremcos.WithLogger(logger),
		gremcos.NumMaxActiveConnections(10),
		gremcos.ConnectionIdleTimeout(time.Second*30),
		gremcos.MetricsPrefix("myservice"),
	)

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create the cosmos connector")
	}

	exitChannel := make(chan struct{})
	go processLoop(cosmos, logger, exitChannel)

	<-exitChannel
	if err := cosmos.Stop(); err != nil {
		logger.Error().Err(err).Msg("Failed to stop cosmos connector")
	}
	logger.Info().Msg("Teared down")
}

func processLoop(cosmos gremcos.Cosmos, logger zerolog.Logger, exitChannel chan<- struct{}) {
	// register for common exit signals (e.g. ctrl-c)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// create tickers for doing health check and queries
	queryTicker := time.NewTicker(time.Second * 2)
	healthCheckTicker := time.NewTicker(time.Second * 30)

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
			err := cosmos.IsHealthy()
			logEvent := logger.Debug()
			if err != nil {
				logEvent = logger.Warn().Err(err)
			}
			logEvent.Bool("healthy", err == nil).Msg("Health Check")
		}
	}

	logger.Info().Msg("Process loop left")
}

func queryCosmos(cosmos gremcos.Cosmos, logger zerolog.Logger) {

	g := api.NewGraph("g")
	// adds an edge from vertex with property name:jan to vertex with property name:hans
	// jan <-knows- hans
	query := g.V().Has("name", "jan").AddE("knows").From(g.V().Has("name", "hans"))
	logger.Info().Msgf("Query: %s", query)

	res, err := cosmos.ExecuteQuery(query)

	if err != nil {
		logger.Error().Err(err).Msg("Failed to execute a gremlin command")
		return
	}

	responses := api.ResponseArray(res)
	values, err := responses.ToValues()
	if err == nil {
		logger.Info().Msgf("Received Values: %v", values)
	}
	properties, err := responses.ToProperties()
	if err == nil {
		logger.Info().Msgf("Received Properties: %v", properties)
	}
	vertices, err := responses.ToVertices()
	if err == nil {
		logger.Info().Msgf("Received Vertices: %v", vertices)
	}
	edges, err := responses.ToEdges()
	if err == nil {
		logger.Info().Msgf("Received Edges: %v", edges)
	}
}
