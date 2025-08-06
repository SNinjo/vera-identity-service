.PHONY: start dev migrate-create migrate-up migrate-down

SHELL := /bin/bash
ENV := source .env &&

start:
	go run main.go

dev:
	reflex -s -r '(\.go$$|^\.env$$)' -- sh -c 'go run cmd/identity/main.go'

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

insert-user-to-db:
	$(ENV) psql "$$DATABASE_URL" -c "INSERT INTO users (email) VALUES ('$(email)');"

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-force:
	$(ENV) migrate -path migrations -database "$$DATABASE_URL" force $(version)
