# Changelog

## v0.1.0 (2020-04-15)

### Bug Fixes

- With [#30](https://github.com/supplyon/gremcos/issues/30) the bug that **each value is url encoded** was fixed.
- With [#29](https://github.com/supplyon/gremcos/issues/29) a bug that could cause a **panic due to closing an already closed channel** was fixed.
- With [#27](https://github.com/supplyon/gremcos/issues/27) the bug that the usage of **characters like '\$' caused the according query to fail** was fixed.
- With [#20](https://github.com/supplyon/gremcos/issues/20) the bug of **concurrent write on the same websocket caused a panic** was fixed.

## Features

- With [#18](https://github.com/supplyon/gremcos/issues/18) **Sonar Cloud integration** for better code quality was added.
- With [#16](https://github.com/supplyon/gremcos/issues/16) **Cosmos DB specific errors** like (`request rate limit exceeded`) **are handled** accordingly. For more information see [ErrorHandling](ErrorHandling.md). Furthermore [prometheus](https://prometheus.io/) **metrics were added** (e.g. request charge, response time and status codes) where added. For more details see [Metrics](Metrics.md).
- With [#12](https://github.com/supplyon/gremcos/issues/12) the **examples were extended**.
- With [#9](https://github.com/supplyon/go-gremlin-cosmos/issues/9) the library is extended in a way that the user does not have to care about connections and connection losses. A **connection pool is integrated to manage connections** (create new on demand and remove expired ones).
- With [#7](https://github.com/supplyon/gremcos/issues/7) https://github.com/ThomasObenaus/go-base/blob/master/health/check.go was implemented for easy **health check of the Cosmos DB connections**.
- With [#6](https://github.com/supplyon/go-gremlin-cosmos/issues/6) **tests where added and test coverage increased**. For being able to test the code some refactorings where necessary.
- With [#3](https://github.com/supplyon/go-gremlin-cosmos/issues/3) **build, test and linting in a local environment and in a Github pipeline** has been implemented. Furthermore the local gremlin server has been fixed to version 3.4.0 and the response format has been tight to Graphson 2.0 (GraphSONMessageSerializerV2d0) in order to be compatible with CosmosDB. Hence all tests where adjusted accordingly. Also the **integration tests were separated from the unit tests** and moved into a separate test suite.
