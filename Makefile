# KEYCLOAK_BIN=/Users/anpks/path/to/keycloak/bin/kc.sh

# inspect: # Atlas migration: to insoect the schema
# 	atlas schema inspect --env gorm

# diff: # Atlas migration: to check the migration difference with DB
# 	atlas migrate diff --env gorm

# apply: # Atlas migration: to apply the migration difference with DB
# 	atlas schema apply --env gorm

# test:  # All tests written will be checked recursively
# 	go test -v -cover ./...

# server:
# 	go run ./cmd/api/main.go

# keycloak-bg: # Start keycloak in dev mode
# #make sure we have updated ~/.zshrc with /path/to/keycloak/bin
# 	nohup kc.sh start-dev --http-host=127.0.0.1 > keycloak.log 2>&1 &


# ----------------------------------------------------------------------------------------- #

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

migrate:
	cd app && go run cmd/migrate/main.go

.PHONY: clean tidy deps migrate swagger run lint test
