BINARY=batcher

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

LDFLAGS=-ldflags "-s -w"

ifeq ($(OS),Windows_NT)
    BINARY := $(BINARY).exe
    RM := del /Q
else
    RM := rm -f
endif

.PHONY: all build test clean run fmt vet lint tidy help docker docker-run

all: test build

download: ## Download project dependencies
	$(GOMOD) download

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY) ./cmd/batcher

test: ## Run tests
	$(GOTEST) -v ./...

clean: ## Clean build files
	$(RM) $(BINARY)
	$(RM) coverage.out

run: build ## Build and run the binary
	./$(BINARY)

fmt: ## Run go fmt
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy up module files
	$(GOMOD) tidy

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker: ## Build docker image
	docker build -t $(BINARY) .

docker-run: docker ## Run docker container
	docker run $(BINARY)

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help