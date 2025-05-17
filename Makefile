# Include .env file
include ./.env
export $(shell sed 's/=.*//' .env)

# Project variables
PROJECT_NAME := search_cdr
PKG := ./...
SERVER := ./cmd/server/
CLIENT := ./cmd/client/
MIGRATIONS_DIR = migrations_prod


# Go commands
BUILD := go build
CLEAN := go clean
FMT := go fmt
VET := go vet
TEST := go test
RUN := go run
MIGRATE := migrate

# Targets
.PHONY: all build clean fmt vet test run migrate-create migrate-up migrate-down migrate-drop

all: fmt vet test build

build-server:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(BUILD) -o ./build/server.exe $(SERVER)

build-client:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  $(BUILD) -o ./build/client.exe $(CLIENT)

clean:
	$(CLEAN)
	rm -f $(PROJECT_NAME)

fmt:
	$(FMT) $(PKG)

vet:
	$(VET) $(PKG)

test:
	$(TEST) $(PKG)


run-server:
	$(RUN) $(SERVER)
run-client:
	$(RUN) $(CLIENT)
