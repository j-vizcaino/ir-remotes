.PHONY: build generate lint test cover

ifneq ($(GO_BUILD_TAGS),)
GO_OPTS := $(GO_OPTS) -tags $(GO_BUILD_TAGS)
endif
COVER := cover.out

build: lint generate
	go build $(GO_OPTS)

generate:
	go generate -v $(GO_OPTS) ./assets/...

lint:
	go vet . ./cmd/... ./pkg/... 
	go fmt . ./cmd/... ./pkg/...

test: lint
	go test -v -coverprofile $(COVER) ./...

cover: test
	go tool cover -html $(COVER)
