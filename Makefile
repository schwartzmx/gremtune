.DEFAULT_GOAL:= all

.PHONY: all
all: vet test

.PHONY: vet
vet:
	@go vet -v

.PHONY: test
test:
	@go test -v

.PHONY: test-bench
test-bench:
	@go test -bench=. -race

.PHONY: gremlin
gremlin:
	@docker build -t gremgo-neptune/gremlin-server -f ./Dockerfile.gremlin .
	@docker run -p 8182:8182 -t gremgo-neptune/gremlin-server