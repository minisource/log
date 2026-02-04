.PHONY: build run test clean docker-build docker-up docker-down migrate

# Build variables
BINARY_NAME=log-service
DOCKER_IMAGE=minisource/log

# Go build
build:
	go build -o bin/$(BINARY_NAME) ./cmd/main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux ./cmd/main.go

# Run locally
run:
	go run ./cmd/main.go

run-hot:
	air -c .air.toml

# Testing
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Linting
lint:
	golangci-lint run

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out

# Dependencies
deps:
	go mod download
	go mod tidy

# Docker
docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

docker-push:
	docker push $(DOCKER_IMAGE):latest

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f log-service

# Development
dev-up:
	docker-compose -f docker-compose.dev.yml up -d

dev-down:
	docker-compose -f docker-compose.dev.yml down

dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f log-service

# Database migrations
migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

# Swagger docs
swagger:
	swag init -g cmd/main.go -o docs

# All-in-one
all: deps lint test build
