# sp-agent

HTTP API base built from the book-hive layout: Gin, Postgres (master/slave), Redis, JWT, request transactions, and Atlas migrations.

## Run

```bash
docker compose up -d postgres redis
go run ./cmd/server
```

Health: `GET /health`. Readiness: `GET /readyz`.

Postgres listens on host port `5440`, Redis on `6380`, so this stack can run next to book-hive.

## Layout

- `cmd/server` — process entry
- `internal/pkg` — config, database, cache, jwt, pagination
- `internal/server` — HTTP server, middleware, validator, and empty controller/repository/service containers
- `schema/schema.sql` — desired schema; generate migrations with `make migrate-diff name=<name>`
