ifneq ("$(wildcard .env)","")
	include .env
	export $(shell sed 's/=.*//' .env)
endif

.PHONY: start
start:
	./bin/main

.PHONY: build
build:
	go build -o=./bin/main ./cmd

.PHONY: test
test:
	go test -v ./...
	
.PHONY: codegen
codegen:
	go run github.com/99designs/gqlgen

.PHONY: proto
proto:
	git submodule update --recursive --remote  
	buf lint
	buf generate --path proto/joyserviceapis

# https://github.com/golang-migrate/migrate

models:
	pg_dump --schema-only url_shortener > schema.sql
	sqlc generate

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path=./migrations -database="$(POSTGRES_URL)" up

migrate-down:
	migrate -path=./migrations -database="$(POSTGRES_URL)" down 1
