.DEFAULT_GOAL := all

all: lint test finish

# This target (taken from: https://gist.github.com/prwhite/8168133) is an easy way to print out a usage/ help of all make targets.
# For all make targets the text after \#\# will be printed.
help: ## Prints the help
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\1\:\2/' | column -c2 -t -s :)"

build: sep ## Builds the library
	@echo "--> Build"
	@go build -race .

test: sep ## Runs all unittests and generates a coverage report.
	@echo "--> Run the unit-tests and checks for race conditions."
	@go test . ./api ./interfaces -timeout 30s -race -run "^Test.*[^IT]$$" -covermode=atomic


test.integration: sep ## Runs all integration tests. As precondition a local gremlin-server has to run and listen on port 8182.
	@echo "--> Run the integration-tests"
	@go test -timeout 30s -run "Test_SuiteIT" -covermode=count

bench: sep ## Execute benchmarks
	@echo "--> Execute benchmarks"
	@go test -race -timeout 30s -v -bench "BenchmarkPoolExecute.*"

lint: ## Runs the linter to check for coding-style issues
	@echo "--> Lint project"
	@echo "!!!!golangci-lint has to be installed. See: https://github.com/golangci/golangci-lint#install"
	@golangci-lint run --fast

report.test: sep ## Runs all unittests and generates a coverage- and a test-report.
	@echo "--> Run the unit-tests"	
	@go test . ./api ./interfaces -timeout 30s -race -run "^Test.*[^IT]$$" -covermode=atomic -coverprofile=coverage.out -json | tee test-report.out

report.lint: ## Runs the linter to check for coding-style issues and generates the report file used in the ci pipeline
	@echo "--> Lint project + Reporting"
	@echo "!!!!golangci-lint has to be installed. See: https://github.com/golangci/golangci-lint#install"
	@golangci-lint run --fast --out-format checkstyle | tee lint.out

gen-mocks: sep ## Generates test doubles (mocks).
	@echo "--> generate mocks (github.com/golang/mock/gomock is required for this)"
	@go install github.com/golang/mock/mockgen@latest
	@mockgen -source=interfaces/websocketConnection.go -destination test/mocks/interfaces/mock_websocketConnection.go
	@mockgen -source=interfaces/dialer.go -destination test/mocks/interfaces/mock_dialer.go
	@mockgen -source=interfaces/queryExecutor.go -destination test/mocks/interfaces/mock_queryExecutor.go
	@mockgen -source=interfaces/queryBuilder.go -destination test/mocks/interfaces/mock_queryBuilder.go
	@mockgen -source=metrics/metrics.go -destination test/mocks/metrics/mock_metrics.go
	@mockgen -source=cosmos.go -destination test/mocks/cosmos/mock_cosmos.go

infra.up: ## Starts up the infra components
	make -C infra up

infra.down: ## Stops up the infra components
	make -C infra down

sep:
	@echo "----------------------------------------------------------------------------------"

finish:
	@echo "=================================================================================="
