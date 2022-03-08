package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	gremcos "github.com/supplyon/gremcos"
	"github.com/supplyon/gremcos/api"
)

type myDynamicCredentialProvider struct {
	credentialFile   string
	UsernameFromFile string `json:"username"`
	PasswordFromFile string `json:"password"`
}

func (dynCred *myDynamicCredentialProvider) updateCredentials() error {
	file, err := os.Open(dynCred.credentialFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&dynCred); err != nil {
		return err
	}

	return nil
}

func (dynCred *myDynamicCredentialProvider) Username() (string, error) {
	if err := dynCred.updateCredentials(); err != nil {
		return "", errors.Wrapf(err, "reading credentials from '%s'", dynCred.credentialFile)
	}

	if len(dynCred.UsernameFromFile) == 0 {
		return "", fmt.Errorf("username not set, use export CDB_USERNAME=/dbs/<cosmosdb name>/colls/<graph name> to specify it")
	}
	return dynCred.UsernameFromFile, nil
}

func (dynCred *myDynamicCredentialProvider) Password() (string, error) {
	if err := dynCred.updateCredentials(); err != nil {
		return "", errors.Wrapf(err, "reading credentials from '%s'", dynCred.credentialFile)
	}

	if len(dynCred.PasswordFromFile) == 0 {
		return "", fmt.Errorf("password not set")
	}
	return dynCred.PasswordFromFile, nil
}

func main() {
	host := os.Getenv("CDB_HOST")
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}).With().Timestamp().Logger()

	if len(host) == 0 {
		logger.Fatal().Msg("Host not set. Use export CDB_HOST=<CosmosDB Gremlin Endpoint> to specify it")
	}
	credentialFile := "./examples/cosmos_dynamic_credentials/credentials.json"
	log.Println("Connecting using:")
	log.Printf("\thost: %s\n", host)
	log.Printf("\tusername: Will be provided by a dynamic credential provder which reads it from the file '%s'\n", credentialFile)
	log.Printf("\tpassword: Will be provided by a dynamic credential provder which reads it from the file '%s'\n", credentialFile)

	credProvider := myDynamicCredentialProvider{credentialFile: credentialFile}
	cosmos, err := gremcos.New(host,
		gremcos.WithResourceTokenAuth(&credProvider),
		gremcos.WithLogger(logger),
		gremcos.NumMaxActiveConnections(10),
		gremcos.ConnectionIdleTimeout(time.Second*30),
		gremcos.MetricsPrefix("myservice"),
		gremcos.AutomaticRetries(3,time.Second * 2),
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
			queryCosmosWithBindings(cosmos, logger)
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

func queryCosmosWithBindings(cosmos gremcos.Cosmos, logger zerolog.Logger) {

	// adds an edge from vertex with property name:jan to vertex with property name:hans
	// jan <-likes- hans
	nameFrom := "jan"
	nameTo := "hans"
	relationship := "likes"

	query := api.NewSimpleQB(`g.V().has("name", nameFrom).addE(relationship).from(g.V().has("name", nameTo))`)
	logger.Info().Msgf("Query: %s", query)

	res, err := cosmos.ExecuteWithBindings(query.String(), map[string]interface{}{
		"nameFrom":     nameFrom,
		"nameTo":       nameTo,
		"relationship": relationship,
	}, nil)

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
