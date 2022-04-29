.PHONY: build test

build:
	go build -o conduit-connector-google-sheets cmd/google-sheets/main.go

test:
	go test $(GOTEST_FLAGS) -race ./...

