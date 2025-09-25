# VidFriends Backend

The VidFriends backend is a Go service that exposes RESTful APIs for authentication,
friend management, and sharing video links enriched with metadata lookups. The
service talks to PostgreSQL via the `pgx` driver and runs database migrations on
startup.

## Project layout

- `cmd/vidfriends/` – entrypoint for running the HTTP server and database migrations.
- `internal/` – packages that contain handlers, data access, and domain logic.
- `migrations/` – SQL migrations applied at startup or via the CLI tooling.

## Running the service

1. Ensure PostgreSQL is running and the connection details in `backend/.env` (copy
   from `.env.example`) are correct.
2. Apply database migrations:
   ```bash
   go run ./cmd/vidfriends migrate up
   ```
3. Start the API server:
   ```bash
   go run ./cmd/vidfriends serve
   ```
   The server listens on `http://localhost:8080` by default.

## Testing

Run the Go test suite:

```bash
go test ./...
```

Use `go test -race ./...` when you need additional data race checks.

## Tooling

- [`air`](https://github.com/cosmtrek/air) or [`reflex`](https://github.com/cespare/reflex)
  can be used for live-reload during development.
- [`golangci-lint`](https://golangci-lint.run/) is recommended to ensure the codebase
  conforms to linting rules before committing changes.
