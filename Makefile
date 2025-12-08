.PHONY: all
all: lint test build

.PHONY: build
build:
	go build -o bin/sleepship

.PHONY: test
test:
	go test ./... -v

.PHONY: test-integration
test-integration:
	go test ./... -tags=integration -v

.PHONY: test-all
test-all:
	go test ./... -tags=integration -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf logs/
	go clean

.PHONY: install
install:
	go install
