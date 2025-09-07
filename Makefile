.PHONY: help install build test clean

help:
	@echo "Available commands:"
	@echo "  make install    - Install all dependencies"
	@echo "  make build      - Build Go daemon"
	@echo "  make test       - Run all tests"
	@echo "  make clean      - Clean build artifacts"

install:
	cd go-daemon && go mod download
	cd python-sdk && pip install -r requirements.txt

build:
	cd go-daemon && go build -o pubsub-daemon .

test:
	cd go-daemon && go test ./...
	cd python-sdk && pytest

clean:
	cd go-daemon && go clean
	rm -f go-daemon/pubsub-daemon
	cd python-sdk && rm -rf build/ dist/ *.egg-info
	find . -type d -name __pycache__ -rm -rf {} +
	find . -type f -name "*.pyc" -delete

run-daemon:
	cd go-daemon && go run .

dev-install:
	cd python-sdk && pip install -e .