BINARY := mtpx
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)

.PHONY: help build fmt vet test cover cover-check lint clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

build: ## Build the binary
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/mtpx

fmt: ## Format sources
	gofmt -l -w .

vet: ## Run go vet
	go vet ./...

test: ## Run tests with race detector and coverage
	go test -race -coverprofile=coverage.out ./...

cover: test ## Print coverage summary
	go tool cover -func=coverage.out

cover-check: ## Fail if the internal packages are below 100% coverage
	go test -covermode=atomic -coverprofile=coverage.internal.out ./internal/...
	@total=$$(go tool cover -func=coverage.internal.out | tail -1 | awk '{print $$3}'); \
	echo "internal coverage: $$total"; \
	if [ "$$total" != "100.0%" ]; then \
		echo "FAIL: internal coverage below 100%"; \
		go tool cover -func=coverage.internal.out | grep -v '100.0%'; \
		exit 1; \
	fi

lint: ## Run golangci-lint
	golangci-lint run

clean: ## Remove build artifacts
	rm -rf bin dist coverage.out
