.PHONY: all fmt vet test test-race cover tidy check

all: check

fmt:
	gofmt -l -w .

vet:
	go vet ./...

test:
	go test ./...

test-race:
	go test -race ./...

cover:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

tidy:
	go mod tidy

check: fmt vet test
