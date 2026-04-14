
.PHONY: test lint cover render


render:
	go run main.go render project-board

lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 || (echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)
	staticcheck ./...

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

cover: test
	go tool cover -html=coverage.out

