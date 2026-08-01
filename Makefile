.PHONY: tidy test api docker-up docker-down migrate-notes

tidy:
	cd backend && go mod tidy

test:
	cd backend && go test ./...

api:
	cd backend && go run ./cmd/api -config configs/config.yaml

api-sqlite:
	cd backend && SFA_DATABASE_DSN='sqlite:file:./sfa_dev.db?cache=shared&mode=rwc' go run ./cmd/api -config configs/config.yaml

docker-up:
	docker compose up -d postgres redis rabbitmq minio

docker-down:
	docker compose down

migrate-notes:
	@echo "P0 uses GORM AutoMigrate on startup. SQL migrations live in backend/migrations for Postgres/prod."
