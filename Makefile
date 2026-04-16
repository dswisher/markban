
.PHONY: test lint cover build browse install uninstall


browse:
	go run . browse project-board

build:
	go build -o markban .

lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 || (echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)
	staticcheck ./...

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

cover: test
	go tool cover -html=coverage.out

install:
	go install .
	@echo "Installed markban to $(shell go env GOPATH)/bin/markban"

uninstall:
	@rm -f "$(shell go env GOPATH)/bin/markban"
	@echo "Uninstalled markban from $(shell go env GOPATH)/bin"

