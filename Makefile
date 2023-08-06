## help: print this help message
.PHONY: help
help: ## This help dialog.
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@read -p "Are you sure? [y/N] " ans; \
	if [ "$${ans:-N}" = "y" ]; then \
		echo "Confirmed. Proceeding..."; \
	else \
		echo "Confirmation canceled. Aborting."; \
		exit 1; \
	fi

# ================================================================================================ #
# DEVELOPMENT
# ================================================================================================ #

## run/api: run the cmd/api application
.PHONY: run/api
run/api: ## Run the app locally
	go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## requirements: Ensures that the go.mod and go.sum files are in sync and that only the required dependencies and
#their correct versions are listed.
.PHONY: requirements
requirements: ## Generate go.mod & go.sum files
	go mod tidy

## db/postgres: Run a postgres container
.PHONY: db/postgres
db/postgres:
	podman run --name=postgres -p 5432:5432 -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
				-e POSTGRES_USER=$POSTGRES_USER -d postgres:15.3-alpine

## db/psql: Interact with the DB using PSQL
.PHONY: db/psql
db/psql:
	 psql ${GREENLIGHT_DB_DSN}

## db/create_db: Create GreenLight Database
.PHONY: db/create_db
db/create_db:
	podman exec -it postgres createdb --username=peter --owner=peter greenlight

## db/drop_db: Drop GreenLight Database
.PHONY: db/drop_db
db/drop_db:
	podman exec -it postgres dropdb --username=peter greenlight

#db/migrations/up: Apply all migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} up

#db/migrations/down: Drop all migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} down

#db/migrations/new: Create new migration file, passing in the name variable
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}
	#make migration name=create_example_table

# ================================================================================================ #
# QUALITY CONTROL
# ================================================================================================ #
## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Formatting code .....'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests......'
	go test -race -vet=off ./...