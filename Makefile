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
    RM := rmdir /s /q
else
    RM := rm -rf
endif

.PHONY: build test clean run fmt vet lint tidy help

download: ## Download project dependencies
	$(GOMOD) download

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o build/$(BINARY) ./cmd/$(BINARY)

test: ## Run tests
	$(GOTEST) ./...

clean: ## Clean build files
	$(RM) build
	$(RM) coverage.out

run: build ## Build and run the binary
	./build/$(BINARY)

fmt: ## Run go fmt
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy up module files
	$(GOMOD) tidy

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Default target
.DEFAULT_GOAL := help
