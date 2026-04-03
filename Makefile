.PHONY: build clean fmt lint test vulncheck mdatagen

build:
	go build

clean:
	go clean

fmt:
	go tool gofumpt -w .

lint:
	go tool staticcheck -checks=all -show-ignored -tests  ./...

test:
	go test ./...

vulncheck:
	go tool govulncheck ./...

mdatagen:
	go tool mdatagen metadata.yaml
