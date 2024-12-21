build: ## Build project with docker and place executable in build directory
	mkdir -p build && CGO_ENABLED=0 GOOS=linux go build -o ./build/server ./cmd/main.go && chmod a+x ./build/server
.PHONY: build

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