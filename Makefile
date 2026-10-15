.PHONY: all build run test clean install

BINARY_NAME=netpulse
BUILD_DIR=bin
CMD_DIR=./cmd/netpulse

all: test build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME) successfully."

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test -v -race ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
