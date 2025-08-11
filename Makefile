.PHONY: wire dev build prod test test-cov test-cov-html access-db migrate-create migrate-up migrate-down migrate-force insert-user-to-db

-include .env
export $(shell sed -n 's/^\([A-Za-z_][A-Za-z0-9_]*\)=.*/\1/p' .env)

wire:
	go run github.com/google/wire/cmd/wire ./internal/app/

dev:
	reflex -s -r '(\.go$$|^\.env$$)' -R '(_gen\.go$$)' -- sh -c 'make wire && go run ./cmd/...'

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/main ./cmd/...
	chmod +x bin/main

prod:
	make build
	GIN_MODE=release bin/main

path ?= ./...
test:
	@cmd="go test $(path)"; \
	if echo "$(MAKECMDGOALS)" | grep -qw verbose; then \
		cmd="$$cmd -v"; \
	fi; \
	if [ -n "$(strip $(name))" ]; then \
		cmd="$$cmd -run $(name)"; \
	fi; \
	echo $$cmd; \
	$$cmd

test-cov:
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out

test-cov-html:
	go tool cover -html=cover.out

access-db:
	psql "$$DATABASE_URL"


MIGRATION_URL := $(DATABASE_URL)&x-migrations-table=$(MIGRATION_TABLE)

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(MIGRATION_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATION_URL)" down 1

migrate-force:
	migrate -path migrations -database "$(MIGRATION_URL)" force $(version)

insert-user-to-db:
	psql "$$DATABASE_URL" -c "INSERT INTO users (email) VALUES ('$(email)');"
