help: ## This help dialog.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

run: ## Run the app locally
	go run ./cmd/api

requirements: ## Generate go.mod & go.sum files
	go mod tidy

postgres:
	podman run --name=postgres -p 5432:5432 -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
				-e POSTGRES_USER=$POSTGRES_USER -d postgres:15.3-alpine

psql:
	 psql ${GREENLIGHT_DB_DSN}

create_db:
	podman exec -it postgres createdb --username=peter --owner=peter greenlight

drop_db:
	podman exec -it postgres dropdb --username=peter greenlight

up:
	@echo 'Running up migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} up

migrate down:
	@echo 'Running down migrations...'
	migrate -path=./migrations -database=${GREENLIGHT_DB_DSN} down