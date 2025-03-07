.PHONY: build test

build: fmt
	cd cmd/aws-overview && go build -o ../../aws-overview

test:
	go test ./...

fmt:
	go fmt ./...
