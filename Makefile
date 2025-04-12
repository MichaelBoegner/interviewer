.PHONY: test test-clean

test:
	@echo "Make: Starting test database..."
	docker compose -f docker-compose.test.yml up -d
	
	@echo "Waiting for Postgres to be ready..."
	@docker compose -f docker-compose.test.yml exec -T test-db bash -c 'for i in {1..10}; do pg_isready -U testuser && exit 0 || sleep 1; done; exit 1'

	@echo "Make: Running migrations..."
	docker run --rm \
		--network="host" \
		-v $(PWD)/database/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database "postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable" \
		up

	@echo "Make: Running tests..."
	@{ \
		go test -v -count=1 ./... ; \
		EXIT_CODE=$$? ; \
		echo "Make: Shutting down test database..." ; \
		docker compose -f docker-compose.test.yml down -v --remove-orphans --timeout 2 || true ; \
		exit $$EXIT_CODE ; \
	}
test-clean:
	@echo "Make: Cleaning up test database..."
	docker compose -f docker-compose.test.yml down -v --remove-orphans --timeout 2 || true