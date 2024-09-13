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

.PHONY: ui	
ui:
	~/go/bin/templ generate

.PHONY: test
test:
	go test -v ./...
	
.PHONY: css
css:
	npx @tailwindcss/cli@next -i styles/tailwind.css -o public/static/css/main.css -m

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
