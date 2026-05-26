SHELL := /bin/zsh

APP_NAME := merger
GO ?= go
CONFIG ?= config/merger.yaml
BUILD_DIR ?= .build
export GOCACHE := $(abspath $(BUILD_DIR)/go-cache)
export GOMODCACHE := $(abspath $(BUILD_DIR)/go-mod-cache)

.PHONY: fmt test build run-ingest run-controlplane compose-up compose-down docker-build clean

$(BUILD_DIR):
	mkdir -p $(GOCACHE) $(GOMODCACHE)

fmt: $(BUILD_DIR)
	$(GO) fmt ./...

test: $(BUILD_DIR)
	$(GO) test ./...

build: $(BUILD_DIR)
	$(GO) build ./cmd/...

run-ingest: $(BUILD_DIR)
	MERGER_CONFIG_PATH=$(CONFIG) $(GO) run ./cmd/merger-ingest

run-controlplane: $(BUILD_DIR)
	MERGER_CONFIG_PATH=$(CONFIG) $(GO) run ./cmd/merger-controlplane

compose-up:
	docker compose -f deployments/local/docker-compose.yml up -d

compose-down:
	docker compose -f deployments/local/docker-compose.yml down

docker-build:
	docker build -f deployments/docker/Dockerfile -t $(APP_NAME):dev .

clean:
	rm -rf ./.build
