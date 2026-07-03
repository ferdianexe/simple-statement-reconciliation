BUILD_DIR      := bin
CLI_BINARY     := reconcile
SERVER_BINARY  := reconcile-server
CLI_DIR        := ./cmd/reconcile
SERVER_DIR     := ./cmd/server

SYS         ?= testdata/system_transactions.csv
BANKS       ?= BCA:testdata/bank_bca.csv,BNI:testdata/bank_bni.csv
START       ?= 2024-01-01
END         ?= 2024-01-31
ADDR        ?= :8080

.PHONY: all build build-cli build-server run serve run-json test test-verbose fmt vet tidy clean

all: build

## build: compile both the CLI and the HTTP server into bin/
build: build-cli build-server

## build-cli: compile the batch CLI into bin/reconcile
build-cli:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(CLI_BINARY) $(CLI_DIR)

## build-server: compile the HTTP server into bin/reconcile-server
build-server:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(SERVER_BINARY) $(SERVER_DIR)

## run: build then run the CLI against the sample testdata (override SYS/BANKS/START/END)
run: build-cli
	./$(BUILD_DIR)/$(CLI_BINARY) -sys $(SYS) -banks "$(BANKS)" -start $(START) -end $(END)

## run-json: same as run, but prints JSON output
run-json: build-cli
	./$(BUILD_DIR)/$(CLI_BINARY) -sys $(SYS) -banks "$(BANKS)" -start $(START) -end $(END) -json

## serve: build then run the HTTP server (override ADDR, default :8080)
serve: build-server
	./$(BUILD_DIR)/$(SERVER_BINARY) -addr $(ADDR)

## test: run the unit test suite
test:
	go test ./...

## test-verbose: run tests with verbose output
test-verbose:
	go test ./... -v

## fmt: format all source files
fmt:
	go fmt ./...

## vet: run go vet static checks
vet:
	go vet ./...

## tidy: sync go.mod/go.sum with actual imports
tidy:
	go mod tidy

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
