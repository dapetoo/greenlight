## help: print this help message
help: ## This help dialog.
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

confirm:
	@read -p "Are you sure? [y/N] " ans; \
	if [ "$${ans:-N}" = "y" ]; then \
		echo "Confirmed. Proceeding..."; \
	else \
		echo "Confirmation canceled. Aborting."; \
		exit 1; \
	fi

## run/api: run the cmd/api application
run/api: ## Run the app locally
	go run ./cmd/api -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

## requirements: Ensures that the go.mod and go.sum files are in sync and that only the required dependencies and their correct versions are listed.
requirements: ## Generate go.mod & go.sum files
	go mod tidy

## db/postgres: Run a postgres container
db/postgres:
	podman run --name=postgres -p 5432:5432 -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
				-e POSTGRES_USER=$POSTGRES_USER -d postgres:15.3-alpine

db/psql:
	 psql ${GREENLIGHT_DB_DSN}

db/create_db:
	podman exec -it postgres createdb --username=peter --owner=peter greenlight

db/drop_db:
	podman exec -it postgres dropdb --username=peter greenlight

db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} up

db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} down

db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}
	#make migration name=create_example_table
