# Changelog

## v?.?.? (2020-??-??)

- With [#9](https://github.com/supplyon/go-gremlin-cosmos/issues/9) the library is extended in a way that the user does not have to care about connections and connection losses. A connection pool is integrated to manage connections (create new on demand and remove expired ones).

- With [#6](https://github.com/supplyon/go-gremlin-cosmos/issues/6) tests where added and test coverage increased. For being able to test the code some refactorings where necessary.

- With [#3](https://github.com/supplyon/go-gremlin-cosmos/issues/3) build, test and linting in a local environment and in a Github pipeline has been implemented. Furthermore the local gremlin server has been fixed to version 3.4.0 and the response format has been tight to Graphson 2.0 (GraphSONMessageSerializerV2d0) in order to be compatible with CosmosDB. Hence all tests where adjusted accordingly. Also the integration tests were separated from the unit tests and moved into a separate test suite.
