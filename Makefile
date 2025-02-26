.PHONY: build test

build:
	cd cmd/aws-overview && go build -o ../../aws-overview

test:
	go test ./...
