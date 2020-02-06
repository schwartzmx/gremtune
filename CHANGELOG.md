# Changelog

## v?.?.? (2020-??-??)

- With [#3](https://github.com/supplyon/go-gremlin-cosmos/issues/3) build, test and linting in a local environment and in a Github pipeline has been implemented. Furthermore the local gremlin server has been fixed to version 3.4.0 and the response format has been tight to Graphson 2.0 (GraphSONMessageSerializerV2d0) in order to be compatible with CosmosDB. Hence all tests where adjusted accordingly. Also the integration tests were separated from the unit tests and moved into a separate test suite.
