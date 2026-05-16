BINARY_DIR := bin

.PHONY: fmt vet test build check clean

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

test:
	go test ./...

build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/ ./cmd/overpatch

check:
	@test -z "$$(gofmt -l ./cmd ./internal)" || (echo "gofmt: unformatted files:"; gofmt -l ./cmd ./internal; exit 1)
	go vet ./...
	go test ./...

clean:
	rm -rf $(BINARY_DIR)
