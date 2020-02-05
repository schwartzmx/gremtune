# Changelog

## v?.?.? (2020-??-??)

- With [#3](https://github.com/supplyon/go-gremlin-cosmos/issues/3) build, test and linting on local environment and in a github pipeline was implemented. Furthermore the local gremlin server was fixed to version 3.4.0 and the response format was tight to Graphson 2.0 (GraphSONMessageSerializerV2d0) in order to be compatible to CosmosDB. Hence all tests where adjusted accordingly. Also the integration tests where separated from the unit tests and moved into a separate test suite.
