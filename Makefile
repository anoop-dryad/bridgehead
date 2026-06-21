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

migrate:
	go run cmd/migrate/main.go

swagger:
	swag init -g cmd/server/main.go -o infra/http/swagger/docs

run: swagger
	go run cmd/server/main.go

.PHONY: migrate swagger run