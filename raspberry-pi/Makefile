build:
	mkdir -p build
	go mod verify
	go mod tidy
	go build -o build/daemon
.PHONY: build

install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.62.2
.PHONY: install-linter

run:
	# run with make run USERNAME=XXXX PASSWORD=XXX
	@echo "Running with username: $(USERNAME) and password: $(PASSWORD)"
	go run main.go --username=$(USERNAME) --password=$(PASSWORD)
.PHONY: run

go-vuln:
	govulncheck ./...
.PHONY: go-vuln

lint:	
	@golangci-lint run
.PHONY: lint