BINARY_NAME=cleat
MAIN_PATH=cmd/cleat/main.go
VERSION?=0.1.11
LDFLAGS=-ldflags "-X github.com/madewithfuture/cleat/internal/cmd.Version=$(VERSION)"

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

clean:
	go clean
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

run: build
	./$(BINARY_NAME)

setup-hooks:
	git config core.hooksPath .githooks

test:
	go test ./...

coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@echo ""
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	threshold=70; \
	echo "Coverage threshold: $${threshold}%"; \
	if awk -v cov="$$coverage" -v thresh="$$threshold" 'BEGIN {if (cov < thresh) exit 1; exit 0}'; then \
		echo "✅ Coverage check passed: $${coverage}% >= $${threshold}%"; \
	else \
		echo "❌ Coverage $${coverage}% is below $${threshold}% threshold"; \
		exit 1; \
	fi

fmt:
	go fmt ./...

vet:
	go vet ./...

install: build
ifeq ($(shell uname), Darwin)
	mkdir -p /usr/local/bin
	cp $(BINARY_NAME) /usr/local/bin/
else
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY_NAME) $(HOME)/.local/bin/
endif

.PHONY: all build build-all clean run install test coverage coverage-html fmt vet setup-hooks
