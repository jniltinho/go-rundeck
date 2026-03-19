APP        := gorundeck
BIN        := bin/$(APP)
VERSION    := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS    := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all dev build css css-watch migrate test lint clean certs build-docker build-docker-prod

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

run:
	./$(BIN) serve

test:
	go test ./...

lint:
	golangci-lint run

certs:
	@echo "Generating SSL certificates..."
	mkdir -p ssl
	openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
		-keyout ssl/server.key -out ssl/server.crt \
		-subj "/C=BR/ST=SP/L=Sao Paulo/O=Development/CN=localhost"

clean:
	rm -rf bin/
	rm -f web/static/css/app.css

build-docker:
	@echo "Building Docker image..."
	docker build --no-cache --progress=plain -t jniltinho/go-rundeck:latest .

build-docker-prod:
	@echo "Building Go application..."
	rm -f $(BIN)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN) .
	upx --best --lzma $(BIN)
