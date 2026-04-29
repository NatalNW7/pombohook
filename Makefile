.PHONY: build-server build-cli build test coverage lint clean

build-server:
	go build -o bin/pombohook-server ./cmd/server/

build-cli:
	go build -o bin/pombo ./cmd/pombo/

build: build-server build-cli

test:
	go test ./... -v -count=1 -race

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out
