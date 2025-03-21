.PHONY: test test-clean

test:
	@echo "Make: Starting test database..."
	docker-compose -f docker-compose.test.yml up -d
	@echo "Make: Running tests..."
	go test ./...
	@echo "Make: Shutting down test database..."
	docker-compose -f docker-compose.test.yml down -v

test-clean:
	@echo "Make: Cleaning up test database..."
	docker-compose -f docker-compose.test.yml down -v