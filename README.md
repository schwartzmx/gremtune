# gremtune

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/schwartzmx/gremtune) [![Build Status](https://travis-ci.org/schwartzmx/gremtune.svg?branch=master)](https://travis-ci.org/schwartzmx/gremtune) [![Go Report Card](https://goreportcard.com/badge/github.com/schwartzmx/gremtune)](https://goreportcard.com/report/github.com/schwartzmx/gremtune)

gremtune is a fork of [qasaur/gremgo](https://github.com/qasaur/gremgo) with alterations to make it compatible with [AWS Neptune](https://aws.amazon.com/neptune/) which is a "Fast, reliable graph database built for the cloud".

gremtune is a fast, efficient, and easy-to-use client for the TinkerPop graph database stack. It is a Gremlin language driver which uses WebSockets to interface with Gremlin Server and has a strong emphasis on concurrency and scalability. Please keep in mind that gremtune is still under heavy development and although effort is being made to fully cover gremtune with reliable tests, bugs may be present in several areas.

**Modifications were made to `gremgo` in order to "support" AWS Neptune's lack of Gremlin-specific features,  like no support query bindings among others. See differences in Gremlin support here: [AWS Neptune Gremlin Implementation Differences](https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-gremlin-differences.html)**

**NOTE**: In order allow continued use for other Graph databases supporting Gremlin outside of only AWS Neptune,  this client still supports querying with Bindings via `ExecuteWithBindings` or `ExecuteFileWithBindings` however this **WILL NOT** work with AWS Neptune as Neptune does not support variables or bindings,

>Variables

>Neptune does not support Gremlin variables and does not support the bindings property.

src: https://docs.aws.amazon.com/neptune/latest/userguide/access-graph-gremlin-differences.html

## Installation
```
go get github.com/schwartzmx/gremtune
dep ensure
```

## Documentation
[GoDoc](https://godoc.org/github.com/schwartzmx/gremtune)

## Examples

#### Single Client
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

    dialer := gremtune.NewDialer("ws://127.0.0.1:8182") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremtune.Dial(dialer, errs) // Returns a gremtune client to interact with
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
#### Pool Client
```go
package main

import (
    "fmt"
    "log"

    "github.com/schwartzmx/gremtune"
)

func main() {
    var gp *Pool
    var gperrs = make(chan error)

	go func(chan error) {
		err := <-gperrs
		log.Fatal("Lost connection to the database: " + err.Error())
    }(gperrs)
    
    dialFn := func() (*Client, error) {
        dialer := gremtune.NewDialer("ws://127.0.0.1:8182")
        c, err := gremtune.Dial(dialer, gperrs)
        if err != nil {
            log.Fatal(err)
        }
        return &c, err
    }
    /* Pool object can be initialized directly

	pool := gremtune.Pool{
		Dial:        dialFn,
		MaxActive:   10,
		IdleTimeout: time.Duration(10 * time.Second),
	}

    or via NewPool() with a PoolConfig */
    pool := gremtune.NewPool(gremtune.PoolConfig{
        Dial: dialFn,
        MaxActive: 10,
        IdleTimeout: time.Duration(10 * time.Second),
    })
    
    res, err := pool.Execute( // Sends a query to Gremlin Server via the Pool
        "g.V('1234')"
    )
    if err != nil {
        fmt.Println(err)
        return
    }
```



#### Example Streaming Query Results
Neptune provides 64 values per Response that is why Execute at present provides a slice of `[]Response` s since it waits for all the responses to be retrieved and then provides it when all have returned.  The `ExecuteAsync` method takes a channel to provide the `Response` as a request parameter and returns the `Response` as and when it is provided by Neptune.  The `Response` are streamed to the caller and once all the `Response`s are provided the channel is closed.
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

    dialer := gremtune.NewDialer("ws://127.0.0.1:8182") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremtune.Dial(dialer, errs) // Returns a gremtune client to interact with
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

## Authentication
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

    dialer := gremtune.NewSecureDialer("127.0.0.1:8182", "username", "password") // Returns a WebSocket dialer to connect to Gremlin Server
    g, err := gremtune.Dial(dialer, errs) // Returns a gremtune client to interact with
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

## License
See [LICENSE](LICENSE.md)
