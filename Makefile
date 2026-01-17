inspect: # Atlas migration: to insoect the schema
	atlas schema inspect --env gorm

diff: # Atlas migration: to check the migration difference with DB
	atlas migrate diff --env gorm

apply: # Atlas migration: to apply the migration difference with DB
	atlas schema apply --env gorm








.PHONY: inspect diff apply