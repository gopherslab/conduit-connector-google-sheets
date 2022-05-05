.PHONY: build test

build:
	go build -o conduit-connector-google-sheets cmd/main.go

test:
	go test $(GOTEST_FLAGS) -race ./...

