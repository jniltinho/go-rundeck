APP        := gorundeck
BIN        := bin/$(APP)
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS    := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all dev build css css-watch migrate test lint clean

all: css build

css:
	tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify

css-watch:
	tailwindcss -i web/static/css/input.css -o web/static/css/app.css --watch

build:
	rm -f $(BIN)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN) .
	upx --best --lzma $(BIN)

dev:
	make css-watch & air -c .air.toml

migrate:
	./$(BIN) migrate

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f web/static/css/app.css
