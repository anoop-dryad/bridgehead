inspect: # Atlas migration: to insoect the schema
	atlas schema inspect --env gorm

diff: # Atlas migration: to check the migration difference with DB
	atlas migrate diff --env gorm

apply: # Atlas migration: to apply the migration difference with DB
	atlas schema apply --env gorm

test:  # All tests written will be checked recursively
	go test -v -cover ./...






.PHONY: inspect diff apply test