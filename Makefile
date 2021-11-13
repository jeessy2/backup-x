.PHONY: build clean test test-race

VERSION=v2.0.0
BIN=backup-x
DIR_SRC=.
DOCKER_CMD=docker

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags="-X main.version=$(VERSION) -X 'main.buildTime=`date`' -extldflags -static"
GO=$(GO_ENV) $(shell which go)
GOROOT=$(shell `which go` env GOROOT)
GOPATH=$(shell `which go` env GOPATH)

build: $(DIR_SRC)/main.go
	@$(GO) build $(GO_FLAGS) -o $(BIN) $(DIR_SRC)

build_image:
	@$(DOCKER_CMD) build -f ./Dockerfile -t backup-x:$(VERSION) .

test:
	@$(GO) test ./... -v

test-race:
	@$(GO) test -race ./...

# clean all build result
clean:
	@$(GO) clean ./...
	@rm -f $(BIN)
	@rm -rf ./dist/*
