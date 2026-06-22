MIGRATE_VERSION := v4.19.1
MIGRATIONS_DIR  := backend/internal/db/migrations

.PHONY: migrate-install
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

.PHONY: migrate-create
migrate-create:
ifndef name
	$(error Usage: make migrate-create name=create_users_table)
endif
	@command -v migrate >/dev/null 2>&1 || { echo "migrate CLI not found — run 'make migrate-install' first."; exit 1; }
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(name)