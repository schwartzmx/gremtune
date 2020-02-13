# gremcos

[![GoDoc](https://godoc.org/github.com/supplyon/gremcos?status.svg)](https://godoc.org/github.com/supplyon/gremcos) ![build](https://github.com/supplyon/gremcos/workflows/build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/supplyon/gremcos)](https://goreportcard.com/report/github.com/supplyon/gremcos)

Gremcos is a fork of [schwartzmx/gremtune](https://github.com/schwartzmx/gremtune) with alterations to make it compatible with [Azure Cosmos](https://docs.microsoft.com/en-us/azure/cosmos-db/introduction) which is a Graph Database (Gremlin API) for Azure.

Gremcos is a fast, efficient, and easy-to-use client for the [TinkerPop](http://tinkerpop.apache.org/docs/current/reference/) graph database stack. It is a gremlin language driver which uses WebSockets to interface with [gremlin server](http://tinkerpop.apache.org/docs/current/reference/#gremlin-server) and has a strong emphasis on concurrency and scalability. Please keep in mind that gremcos is still under heavy development and although effort is being made to fully cover gremcos with reliable tests, bugs may be present in several areas.

## Installation

```bash
go get github.com/supplyon/gremcos
```

## Examples

- See: [examples/README.md](examples/README.md)

## Hints

This implementation supports [Graphson 2.0](http://tinkerpop.apache.org/docs/3.4.4/dev/io/#graphson-2d0) (not 3) in order to be compatible to CosmosDB. This means all the responses from the CosmosDB server as well as the responses from the local gremlin-server have to comply with the 2.0 format.

## License

See [LICENSE](LICENSE.md)

### 3rd Party Licenses

- [difflib license](https://github.com/pmezard/go-difflib/blob/master/LICENSE)
