.PHONY: build clean fmt lint test

GO_TOOLS := -modfile=tools.mod

build:
	go build

clean:
	go clean

fmt:
	go tool $(GO_TOOLS) gofumpt -w .

lint:
	go tool $(GO_TOOLS) staticcheck -checks=all ./...

test:
	go test ./...
