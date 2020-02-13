# gremcos

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/schwartzmx/gremtune) [![Build Status](https://travis-ci.org/schwartzmx/gremtune.svg?branch=master)](https://travis-ci.org/schwartzmx/gremtune) [![Go Report Card](https://goreportcard.com/badge/github.com/schwartzmx/gremtune)](https://goreportcard.com/report/github.com/schwartzmx/gremtune)

- [ ] **Adjust README --> go-gremlin-cosmos**

gremtune is a fork of [qasaur/gremgo](https://github.com/qasaur/gremgo) with alterations to make it compatible with [AWS Neptune](https://aws.amazon.com/neptune/) which is a "Fast, reliable graph database built for the cloud".

gremtune is a fast, efficient, and easy-to-use client for the TinkerPop graph database stack. It is a Gremlin language driver which uses WebSockets to interface with Gremlin Server and has a strong emphasis on concurrency and scalability. Please keep in mind that gremtune is still under heavy development and although effort is being made to fully cover gremtune with reliable tests, bugs may be present in several areas.

**Modifications were made to `gremgo` in order to "support" AWS Neptune's lack of Gremlin-specific features, like no support query bindings among others. See differences in Gremlin support here: [AWS Neptune Gremlin Implementation Differences](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-gremlin-differences.html)**

## Installation

```
go get github.com/schwartzmx/gremtune
dep ensure
```

## Documentation

- [GoDoc](https://godoc.org/github.com/schwartzmx/gremtune)

## Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/schwartzmx/gremtune"
)

func main() {
    errs := make(chan error)
    go func(chan error) {
        err := <-errs
        log.Fatal("Lost connection to the database: " + err.Error())
    }(errs) // Example of connection error handling logic

    dialer := gremcos.NewDialer("ws://127.0.0.1:8182") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremcos.Dial(dialer, errs) // Returns a gremcos client to interact with
    if err != nil {
        fmt.Println(err)
        return
    }
    res, err := g.Execute( // Sends a query to Gremlin Server
        "g.V('1234')"
    )
    if err != nil {
        fmt.Println(err)
        return
    }
    j, err := json.Marshal(res[0].Result.Data) // res will return a list of resultsets,  where the data is a json.RawMessage
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    fmt.Printf("%s", j)
}
```

### Example for streaming the result

Neptune provides 64 values per Response that is why Execute at present provides a [] of Response since it waits for all the responses to be retrieved and then provides it.In ExecuteAsync method it takes a channel to provide the Response as request parameter and provides the Response as and when it is provided by Neptune. The Response are streamed to the caller and once all the Responses are provided the channel is closed.
go test -v -run ExecuteBulkDataAsync is the cmd to run the testcase)

```go
package main

import (
    "fmt"
    "log"
    "time"
    "strings"
    "github.com/schwartzmx/gremtune"
)

func main() {
    errs := make(chan error)
    go func(chan error) {
        err := <-errs
        log.Fatal("Lost connection to the database: " + err.Error())
    }(errs) // Example of connection error handling logic

    dialer := gremcos.NewDialer("ws://127.0.0.1:8182") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremcos.Dial(dialer, errs) // Returns a gremcos client to interact with
    if err != nil {
        fmt.Println(err)
        return
    }
    start := time.Now()
    responseChannel := make(chan AsyncResponse, 10)
    err := g.ExecuteAsync( // Sends a query to Gremlin Server
        "g.V().hasLabel('Employee').valueMap(true)", responseChannel
    )
    log.Println(fmt.Sprintf("Time it took to execute query %s", time.Since(start)))
    if err != nil {
        fmt.Println(err)
        return
    }
    count := 0
    asyncResponse := AsyncResponse{}
    start = time.Now()
    for asyncResponse = range responseChannel {
        log.Println(fmt.Sprintf("Time it took to get async response: %s response status: %v", time.Since(start), asyncResponse.Response.Status.Code))
        count++

        nl := new(BulkResponse)
        datastr := strings.Replace(string(asyncResponse.Response.Result.Data), "@type", "type", -1)
        datastr = strings.Replace(datastr, "@value", "value", -1)
        err = json.Unmarshal([]byte(datastr), &nl)
        if err != nil {
           fmt.Println(err)
           return nil, err
        }
        log.Println(fmt.Sprintf("No of rows retrieved: %v", len(nl.Value)))
        start = time.Now()
    }
}
```

### Authentication

The plugin accepts authentication creating a secure dialer where credentials are setted.
If the server where are you trying to connect needs authentication and you do not provide the
credentials the complement will panic.

```go
package main

import (
    "fmt"
    "log"

    "github.com/schwartzmx/gremtune"
)

func main() {
    errs := make(chan error)
    go func(chan error) {
        err := <-errs
        log.Fatal("Lost connection to the database: " + err.Error())
    }(errs) // Example of connection error handling logic

    dialer := gremcos.NewSecureDialer("127.0.0.1:8182", "username", "password") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremcos.Dial(dialer, errs) // Returns a gremcos client to interact with
    if err != nil {
        fmt.Println(err)
        return
    }
    res, err := g.Execute( // Sends a query to Gremlin Server
        "g.V('1234')"
    )
    if err != nil {
        fmt.Println(err)
        return
    }
    j, err := json.Marshal(res[0].Result.Data) // res will return a list of resultsets,  where the data is a json.RawMessage
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    fmt.Printf("%s", j)
}
```

## Hints

This implementation supports [Graphson 2.0](http://tinkerpop.apache.org/docs/3.4.4/dev/io/#graphson-2d0) (not 3) in order to be compatible to CosmosDB. This means all the responses from the CosmosDB server as well as the responses from the local gremlin-server have to comply with the 2.0 format.

## Examples

Examples can be found at [examples/](examples/).
To run it on can use: `go run ./examples`

## License

See [LICENSE](LICENSE.md)

### 3rd Party Licenses

- [difflib license](https://github.com/pmezard/go-difflib/blob/master/LICENSE)
