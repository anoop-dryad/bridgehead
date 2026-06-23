clean:
	cd app && go clean -modcache

tidy:
	cd app && go mod tidy

deps:
	cd app && go mod download

swagger:
	cd app && swag init -g cmd/server/main.go -o infra/http/swagger/docs

run: swagger
	cd app && go run cmd/server/main.go

lint:
	cd app && golangci-lint run ./...

test:
	cd app && go test ./...

vuln:
	cd app && govulncheck ./...

migrate:
	cd app && go run cmd/migrate/main.go

# ── Docker ───────────────────────────────────────────
docker-build: # builds server image
	docker build -f app/Dockerfile -t bridgehead:local app/

docker-migrate: # builds migration image
	docker build -f app/Dockerfile.migrate -t bridgehead-migrate:local app/

docker-run: # runs migrations in container
	docker run --env-file .env -p 8080:8080 bridgehead:local

docker-migrate-run: # runs server in container
	docker run --env-file .env bridgehead-migrate:local

.PHONY: clean tidy deps migrate swagger run lint test vuln docker-build docker-migrate docker-run docker-migrate-run
