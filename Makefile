
.PHONY: test lint cover render


# Note that this builds a binary and runs it, rather than just using "go run". That is because a
# binary with a stable name only pops up the "Allow or Deny" message once.
render:
	go build -o .bin/markban . && .bin/markban render project-board

lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 || (echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)
	staticcheck ./...

test:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total

cover: test
	go tool cover -html=coverage.out

