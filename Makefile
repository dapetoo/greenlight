postgres:
	podman run --name=postgres -p 5432:5432 -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
				-e POSTGRES_USER=$POSTGRES_USER -d postgres:15.3-alpine

psql:
	 psql $GREENLIGHT_DB_DSN

create_db:
	podman exec -it postgres createdb --username=peter --owner=peter greenlight

drop_db:
	podman exec -it postgres dropdb --username=peter greenlight

migrate:
	migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up

migrate down:
	migrate -path=./migrations -database=$GREENLIGHT_DB_DSN down