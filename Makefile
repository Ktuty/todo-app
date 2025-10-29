# Build
build:
	docker-compose build todo-app

# Run
run:
	docker-compose up todo-app

# Tests
test:
	go test -v ./...

test-unit:
	go test -v ./pkg/... -short

# Database
migrate:
	migrate -path ./schema -database 'postgres://postgres:Taran@12092004@localhost:5432/test?sslmode=disable' up

migrate-down:
	migrate -path ./schema -database 'postgres://postgres:Taran@12092004@localhost:5432/test?sslmode=disable' down

# Documentation
swag:
	swag init -g cmd/main.go

# RabbitMQ Tests (Windows compatible)
all-rabbit-tests:
	@make test-rabbitmq-http
	@make test-load
	@make test-auth
	@make test-rabbitmq-create-user
	@make test-rabbitmq-actions
	@make test-rabbitmq-stats
	@make health-full


test-rabbitmq-create-user:
	@echo "Testing RabbitMQ create_user action..."
	@go run cmd/test_rabbitmq/client.go

test-rabbitmq-actions:
	@echo "Testing all RabbitMQ actions..."
	@go run cmd/test_rabbitmq/test_other_actions.go

test-rabbitmq-http:
	@echo "Testing RabbitMQ HTTP endpoints..."
	@echo.
	@echo "=== Health Check ==="
	@curl -s http://localhost:8000/rabbitmq/health
	@echo.
	@echo.
	@echo "=== Stats ==="
	@curl -s http://localhost:8000/rabbitmq/stats
	@echo.
	@echo.
	@echo "=== Send Test Message ==="
	@curl -s -X POST http://localhost:8000/rabbitmq/send -H "Content-Type: application/json" -d "{\"id\": \"make-test-001\", \"version\": \"v1\", \"action\": \"health_check\", \"data\": {}, \"auth\": \"todo-app-api-key-12345\"}"

test-rabbitmq-stats:
	@echo "Checking RabbitMQ statistics..."
	@curl -s http://localhost:8000/rabbitmq/stats

# Infrastructure
infra-up:
	docker-compose up -d rabbitmq

infra-down:
	docker-compose down

infra-logs:
	docker-compose logs -f rabbitmq

infra-status:
	@echo "RabbitMQ Management: http://localhost:15672 (admin:password)"
	@curl -s http://localhost:15672/api/healthchecks/node -u admin:password > nul 2>&1 && echo "RabbitMQ is healthy" || echo "RabbitMQ management interface not accessible"

# Development
dev:
	@echo "Starting development environment..."
	@make infra-up
	@ping -n 5 127.0.0.1 > nul
	@go run cmd/main.go

dev-clean:
	@echo "Cleaning development environment..."
	@docker-compose down -v
	@docker system prune -f

# Health Checks (Windows compatible)
health:
	@echo "Checking service health..."
	@echo.
	@echo "=== Main Health ==="
	@curl -s http://localhost:8000/health
	@echo.
	@echo.
	@echo "=== RabbitMQ Health ==="
	@curl -s http://localhost:8000/rabbitmq/health

health-full:
	@make health
	@echo.
	@echo "=== RabbitMQ Stats ==="
	@curl -s http://localhost:8000/rabbitmq/stats

# Security Tests (Windows compatible)
test-auth:
	@echo "Testing authentication..."
	@echo.
	@echo "Testing with valid API key..."
	@curl -s -X POST http://localhost:8000/rabbitmq/send -H "Content-Type: application/json" -d "{\"id\": \"auth-test-1\", \"version\": \"v1\", \"action\": \"health_check\", \"data\": {}, \"auth\": \"todo-app-api-key-12345\"}"
	@echo.
	@echo.
	@echo "Testing with invalid API key..."
	@curl -s -X POST http://localhost:8000/rabbitmq/send -H "Content-Type: application/json" -d "{\"id\": \"auth-test-2\", \"version\": \"v1\", \"action\": \"health_check\", \"data\": {}, \"auth\": \"invalid-key\"}"


# Installation helpers
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully!"

# Quick Start
quick-start:
	@echo "🚀 Quick Start: Setting up Todo App with RabbitMQ"
	@echo "1. Starting infrastructure..."
	@make infra-up
	@echo "2. Waiting for RabbitMQ to be ready..."
	@ping -n 10 127.0.0.1 > nul
	@echo "3. Starting application..."
	@go run cmd/main.go

# Clean
clean:
	@go clean

# Help
help:
	@echo "Available commands:"
	@echo "  build              - Build Docker image"
	@echo "  run                - Run application with Docker"
	@echo "  dev                - Start development environment"
	@echo "  quick-start        - Quick setup and start"
	@echo.
	@echo "Testing:"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-rabbitmq-http   - Test RabbitMQ HTTP endpoints"
	@echo "  test-load          - Run load test"
	@echo "  test-auth          - Test authentication"
	@echo.
	@echo "Infrastructure:"
	@echo "  infra-up           - Start RabbitMQ"
	@echo "  infra-down         - Stop infrastructure"
	@echo "  infra-status       - Check infrastructure status"
	@echo.
	@echo "Database:"
	@echo "  migrate            - Run database migrations"
	@echo "  migrate-down       - Rollback migrations"
	@echo.
	@echo "Code Quality:"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  tidy               - Tidy dependencies"
	@echo.
	@echo "Health Checks:"
	@echo "  health             - Check service health"
	@echo "  health-full        - Full health check"
	@echo.
	@echo "Documentation:"
	@echo "  swag               - Generate Swagger docs"
	@echo.
	@echo "Helpers:"
	@echo "  clean              - Clean build artifacts"
	@echo "  help               - Show this help"

.PHONY: build run test test-unit migrate migrate-down swag \
        test-rabbitmq-create-user test-rabbitmq-actions \
        test-rabbitmq-http test-rabbitmq-stats test-rabbitmq-simple \
        infra-up infra-down infra-logs infra-status dev dev-clean \
        health health-full test-load test-auth fmt vet tidy \
        install-tools test-all clean help quick-start