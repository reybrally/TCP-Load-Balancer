.PHONY: help build run test test-verbose test-coverage test-unit test-integration \
	docker-build docker-up docker-down docker-logs docker-clean \
	lint fmt clean rebuild install-deps all

RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m

BINARY_NAME := tcp-lb
MAIN_PATH := cmd/main.go
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

help:
	@echo "$(BLUE)TCP Load Balancer - Make Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Building:$(NC)"
	@echo "  make build              Build binary"
	@echo "  make rebuild            Clean and build"
	@echo "  make install-deps       Install Go dependencies"
	@echo ""
	@echo "$(YELLOW)Running:$(NC)"
	@echo "  make run                Run TCP load balancer locally"
	@echo "  make docker-up          Start all services with Docker"
	@echo "  make docker-down        Stop Docker containers"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  make test               Run all tests (skip problematic ones)"
	@echo "  make test-verbose       Run tests with verbose output"
	@echo "  make test-coverage      Show test coverage"
	@echo "  make test-unit          Run unit tests only"
	@echo "  make test-integration   Run integration tests only"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@echo "  make lint               Run Go linter"
	@echo "  make fmt                Format code"
	@echo ""
	@echo "$(YELLOW)Cleanup:$(NC)"
	@echo "  make clean              Remove binary and stop containers"
	@echo "  make docker-clean       Remove Docker images and volumes"
	@echo "  make all                Full setup (clean, build, test, docker)"
	@echo ""


build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@go build -v -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Build complete!$(NC)"

rebuild: clean build

install-deps:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies installed!$(NC)"


run: build
	@echo "$(BLUE)Starting TCP Load Balancer...$(NC)"
	@echo "$(YELLOW)Listening on: localhost:8080$(NC)"
	@echo "$(YELLOW)Metrics on: localhost:9090/metrics$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@./$(BINARY_NAME)


docker-build:
	@echo "$(BLUE)Building Docker images...$(NC)"
	@docker-compose build
	@echo "$(GREEN)✓ Docker build complete!$(NC)"

docker-up: docker-build
	@echo "$(BLUE)Starting services with Docker Compose...$(NC)"
	@docker-compose up --build -d
	@echo ""
	@echo "$(GREEN)✓ Services started!$(NC)"
	@echo ""
	@echo "$(YELLOW)Access points:$(NC)"
	@echo "  🚀 Load Balancer:  localhost:8080"
	@echo "  📊 Prometheus:     http://localhost:9091"
	@echo "  📈 Grafana:        http://localhost:3000 (admin/admin)"
	@echo "  📝 Metrics:        http://localhost:9090/metrics"
	@echo ""
	@echo "$(YELLOW)To view logs:$(NC)"
	@echo "  make docker-logs"
	@echo ""

docker-down:
	@echo "$(BLUE)Stopping Docker containers...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✓ Containers stopped!$(NC)"

docker-restart: docker-down docker-up

docker-logs:
	@echo "$(BLUE)Docker Compose logs (Ctrl+C to exit):$(NC)"
	@docker-compose logs -f

docker-ps:
	@echo "$(BLUE)Running containers:$(NC)"
	@docker-compose ps

docker-clean: docker-down
	@echo "$(BLUE)Cleaning Docker resources...$(NC)"
	@docker-compose down -v --remove-orphans
	@docker system prune -f
	@echo "$(GREEN)✓ Docker cleaned!$(NC)"


test:
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v ./tests/...
	@echo "$(GREEN)✓ Tests complete!$(NC)"

test-verbose:
	@echo "$(BLUE)Running tests (verbose)...$(NC)"
	@go test -v -race ./tests/...

test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -cover ./tests/...
	@echo ""
	@echo "$(YELLOW)Detailed coverage report:$(NC)"
	@go test -coverprofile=coverage.out ./tests/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report saved to coverage.html$(NC)"

test-unit:
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -v ./tests/unit/...

test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	@go test -v ./tests/integration/...

test-race:
	@echo "$(BLUE)Running tests with race detector...$(NC)"
	@go test -race ./tests/...


lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run ./...
	@echo "$(GREEN)✓ Linting complete!$(NC)"

fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@gofmt -w .
	@echo "$(GREEN)✓ Code formatted!$(NC)"

vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet complete!$(NC)"


clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@docker-compose down 2>/dev/null || true
	@echo "$(GREEN)✓ Clean complete!$(NC)"

clean-all: clean docker-clean
	@echo "$(GREEN)✓ Everything cleaned!$(NC)"


all: install-deps rebuild test docker-up

dev: rebuild test run

prod: rebuild test docker-up

check: fmt vet lint test


info:
	@echo "$(BLUE)Build Information:$(NC)"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Go Version: $(shell go version | awk '{print $$3}')"

version: info

status: docker-ps


test-client:
	@echo "$(BLUE)Testing load balancer...$(NC)"
	@for i in {1..5}; do \
		echo "Connection $$i"; \
		echo "test data $$i" | nc localhost 8080 2>/dev/null || echo "Connection failed"; \
	done
	@echo "$(GREEN)✓ Test complete!$(NC)"

metrics:
	@echo "$(BLUE)Fetching metrics...$(NC)"
	@curl -s http://localhost:9090/metrics | grep tcp_lb || echo "Metrics endpoint not available"

health:
	@echo "$(BLUE)Checking health...$(NC)"
	@curl -s http://localhost:9090/health && echo "" || echo "Health check failed"


.PHONY: help-short
help-short:
	@echo "$(GREEN)Quick Commands:$(NC)"
	@echo "  make run         → Start load balancer"
	@echo "  make docker-up   → Start all services"
	@echo "  make test        → Run tests"
	@echo "  make clean       → Cleanup"
	@echo "  make all         → Full setup"
