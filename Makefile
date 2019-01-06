.PHONY: build lint test cover

COVER := cover.out

build: lint
	go build

lint:
	go vet . ./cmd/... ./pkg/... 
	go fmt . ./cmd/... ./pkg/...

test: lint
	go test -v -coverprofile $(COVER) ./...

cover: test
	go tool cover -html $(COVER)
