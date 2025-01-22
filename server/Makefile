build-server:
	mkdir -p build
	go mod verify
	go mod tidy
	go build -o build/server
.PHONY: build-server

install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2
.PHONY: install-linter

run:
	go run main.go
.PHONY: run

go-vuln:
	govulncheck ./...
.PHONY: go-vuln

lint:	
	@golangci-lint run
.PHONY: lint

proto: ## Builds proto files
	@../proto.sh
.PHONY: proto

test: ## Starts tests locally
	go test ./... --count=1 -p 1 # -p 1 do not run tests parallel. Avoid race conditions when starting/tearing down tests
.PHONY: test

build: proto build-server