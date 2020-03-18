package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	gremcos "github.com/supplyon/gremcos"
	"github.com/supplyon/gremcos/api"
	"github.com/supplyon/gremcos/interfaces"
)

func main() {

	host := "localhost"
	port := 8182
	hostURL := fmt.Sprintf("ws://%s:%d/gremlin", host, port)
	logger := zerolog.New(os.Stdout).Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat}).With().Timestamp().Logger()

	cosmos, err := gremcos.New(hostURL, gremcos.WithLogger(logger), gremcos.NumMaxActiveConnections(10), gremcos.ConnectionIdleTimeout(time.Second*1))
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

func processLoop(cosmos *gremcos.Cosmos, logger zerolog.Logger, exitChannel chan<- struct{}) {
	// register for common exit signals (e.g. ctrl-c)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// create tickers for doing health check and queries
	queryTicker := time.NewTicker(time.Millisecond * 20)
	healthCheckTicker := time.NewTicker(time.Second * 20)

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
			os.Exit(1)
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

func queryCosmos(cosmos *gremcos.Cosmos, logger zerolog.Logger) {

	// values
	// []TypedValue
	data := `["bla",true,{"@type":"g:Int32","@value":1287}]`
	values, err := api.ToValues(json.RawMessage(data))
	if err != nil {
		panic(err)
	}
	fmt.Println("\n###### VALUES ##########################################")
	fmt.Printf("IN: %s\n", data)
	fmt.Printf("OUT: %s\n", values)

	// properties
	// type, {id, label, value}
	data = `[{
	"@type":"g:VertexProperty",
	"@value":{
	"id":{
		"@type":"g:Int64",
		"@value":30
	},
	"value":"prop value",
	"label":"prop key"
	}
}]`

	properties, err := api.ToProperties(json.RawMessage(data))
	if err != nil {
		panic(err)
	}

	fmt.Println("\n###### PROPERTIES ##########################################")
	fmt.Printf("IN: %s\n", data)
	fmt.Printf("OUT: %s\n", properties)

	// valueMap
	// map[string]TypedValue
	data = `[{
	"string":["bla"],
	"bool":[true],
	"int":[{"@type":"g:Int32","@value":1287}]
}]`
	valueMap, err := api.ToValueMap(json.RawMessage(data))
	if err != nil {
		panic(err)
	}

	fmt.Println("\n###### VALUEMAP ##########################################")
	fmt.Printf("IN: %s\n", data)
	fmt.Printf("OUT: %s\n", valueMap)

	// vertex
	// type, {id, label}
	data = `[{
	"@type":"g:Vertex",
	"@value":{
	"id":{
		"@type":"g:Int64",
		"@value":29
	},
	"label":"user"
	}
}]`
	vertex, err := api.ToVertex(json.RawMessage(data))
	if err != nil {
		panic(err)
	}

	fmt.Println("\n###### VERTEX ##########################################")
	fmt.Printf("IN: %s\n", data)
	fmt.Printf("OUT: %s\n", vertex)

	// Edge
	// type, {id, label, inVLabel,outVLabel,{id,id}}
	data = `[{
	"@type":"g:Edge",
	"@value":{
		"id":{
			"@type":"g:Int64",
			"@value":38
		},
		"label":"knows",
		"inVLabel":"user",
		"outVLabel":"user",
		"inV":{
			"@type":"g:Int64",
			"@value":29
		},
		"outV":{
			"@type":"g:Int64",
			"@value":33
		}
	}
}]`
	edge, err := api.ToEdge(json.RawMessage(data))
	if err != nil {
		panic(err)
	}
	fmt.Println("\n###### Edge ##########################################")
	fmt.Printf("IN: %s\n", data)
	fmt.Printf("OUT: %s\n", edge)

	if true {
		return
	}

	g := api.NewGraph("g")
	query := g.AddV("User").Property("userid", "12345").Property("email", "max.mustermann@example.com").Id()
	query = g.VBy(33)
	query = g.VBy(29)
	logger.Info().Msgf("Query: %s", query)
	res, err := cosmos.ExecuteQuery(query)
	queryStr := "g.addE('knows').from(g.V(33)).to(g.V(29))"
	res, err = cosmos.Execute(queryStr)

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

	//vert, err := api.ToProperties(res)
	//if err != nil {
	//	logger.Error().Err(err).Msg("Failed to map the response to a vertex")
	//}
	//
	//logger.Info().Msgf("Vertex: %v", vert)
	//spew.Dump(vert)
}

func queryCosmosAsync(cosmos *gremcos.Cosmos, logger zerolog.Logger) {
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
