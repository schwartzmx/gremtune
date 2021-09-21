package main

import (
	"log"
	"os"
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
		gremcos.WithAuth(username, password), // <- static password obtained and set only once at startup
		gremcos.WithLogger(logger),
		gremcos.NumMaxActiveConnections(10),
		gremcos.ConnectionIdleTimeout(time.Second*30),
		gremcos.MetricsPrefix("myservice"),
	)

	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create the cosmos connector")
	}

	queryCosmos(cosmos, logger)
	queryCosmosWithBindings(cosmos, logger)

	if err := cosmos.Stop(); err != nil {
		logger.Error().Err(err).Msg("Failed to stop cosmos connector")
	}
	logger.Info().Msg("Teared down")
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
