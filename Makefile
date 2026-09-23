GO        ?= go
GOFLAGS   ?= -trimpath
LDFLAGS   := -s -w -X main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)
COVER     ?= coverage.out

.PHONY: all build agent controller test lint fmt tidy migrate-up migrate-down web docker clean

all: build

build: agent controller

agent:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/agent ./agent/cmd/agent

controller:
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o bin/controller ./controller/cmd/controller

test:
	$(GO) test ./... -count=1 -coverprofile=$(COVER)

test-race:
	$(GO) test ./... -count=1 -race

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	$(GO) mod tidy

migrate-up:
	$(GO) run ./controller/cmd/migrate up

migrate-down:
	$(GO) run ./controller/cmd/migrate down

web:
	cd web && npm install && npm run build

docker:
	docker compose -f deploy/docker-compose.yml up -d --build

clean:
	rm -rf bin/ $(COVER)
