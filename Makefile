.PHONY: test

GOOS ?=
GOARCH ?=

all: bin/envexpander

bin/envexpander: *.go cmd/envexpander/*.go providers/*.go
	go build -o $@ ./cmd/envexpander

test:
	go test -race -cover $(shell go list ./... | grep -v /vendor/ | grep -v /mocks)
