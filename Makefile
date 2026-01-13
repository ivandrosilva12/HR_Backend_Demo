.PHONY: up migrate_up migrate_down status down logs test swag

up:
	@echo "📦 Subindo containers..."
	docker compose --env-file .env up -d --build

status:
	@echo "📊 Verificando status dos containers..."
	docker-compose ps

migrate_up:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/hrms?sslmode=disable" up

migrate_down:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/hrms?sslmode=disable" down

logs:
	@echo "📝 Logs da aplicação..."
	docker-compose logs -f api

down:
	@echo "🧹 Parando e removendo containers..."
	docker-compose down

test:
	@echo "🧪 Executando testes unitários..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

swag:
	swag init --parseDependency --parseInternal --output docs