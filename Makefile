.PHONY: wire dev build prod test test-cov test-cov-html access-db migrate-create migrate-up migrate-down migrate-force insert-user-to-db

SHELL := /bin/bash
ENV := source .env &&

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
	$(ENV) GIN_MODE=release bin/main

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
	$(ENV) psql "$$DATABASE_URL"

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-force:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" force $(version)

insert-user-to-db:
	$(ENV) psql "$$DATABASE_URL" -c "INSERT INTO users (email) VALUES ('$(email)');"
