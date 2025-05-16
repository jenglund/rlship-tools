.PHONY: test test-backend test-frontend dev dev-backend dev-frontend clean

# Default target
all: test

# Testing targets
test: test-backend test-frontend

test-backend:
	@echo "Running backend tests..."
	cd backend && go test -v -race -coverprofile=coverage.out ./...
	cd backend && go tool cover -func=coverage.out

test-frontend:
	@echo "Running frontend tests..."
	cd apps/mobile && npm test -- --coverage

# Development targets
dev: dev-backend dev-frontend

dev-backend:
	@echo "Starting backend services..."
	docker-compose up -d postgres redis
	cd backend && go run cmd/api/main.go

dev-frontend:
	@echo "Starting frontend development server..."
	cd apps/mobile && npm run web

# Clean up
clean:
	@echo "Cleaning up..."
	docker-compose down
	rm -f backend/coverage.out
	find . -name "node_modules" -type d -prune -exec rm -rf '{}' + 