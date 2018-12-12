.PHONY: build lint test

build: lint
	go build

lint:
	go vet . ./cmd/... ./pkg/... 
	go fmt . ./cmd/... ./pkg/...

test: lint
	go test -v ./cmd/... ./pkg/...	
