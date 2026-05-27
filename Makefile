
.PHONY: test lint cover build serve install uninstall


serve:
	go run . serve project-board

build:
	go build -o markban .

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

cover: test
	go tool cover -html=coverage.out

install:
	go install .
	@echo "Installed markban to $(shell go env GOPATH)/bin/markban"

uninstall:
	@if [ -f "$(shell go env GOPATH)/bin/markban" ]; then \
		rm -f "$(shell go env GOPATH)/bin/markban"; \
		echo "Uninstalled markban from $(shell go env GOPATH)/bin"; \
	else \
		echo "markban not installed"; \
	fi

