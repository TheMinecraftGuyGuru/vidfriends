# VidFriends Contributor Guide

Welcome to the VidFriends project! This guide walks contributors through the day-to-day workflows for preparing a local environment, running database migrations, seeding useful development data, and executing the full automated test suite.

## Prerequisites

Before diving in ensure the following tooling is available:

- **PostgreSQL 15+** – locally installed or via Docker Compose using the files under `deploy/`.
- **Go 1.21+** – required for backend development and running the migration CLI.
- **Node.js 18+ with pnpm** – used by the Vite-based frontend and its test harness.
- **yt-dlp** – optional locally; the Docker Compose stack provides it automatically for metadata and video downloads.

Copy the example environment files the first time you set up the project:

```bash
cp configs/backend.env.example backend/.env
cp configs/frontend.env.example frontend/.env
cp configs/deploy.env.example deploy/.env # only if you are using docker compose
```

Update credentials and secrets inside the copied files. The backend `.env` must include `VIDFRIENDS_DATABASE_URL` that points at the database you plan to use for development/testing.

## Running database migrations

The backend exposes a light wrapper around our migration tooling. From the `backend/` directory run:

```bash
cd backend
go run ./cmd/vidfriends migrate up
```

The command reads configuration from `backend/.env` (via the `VIDFRIENDS_` environment variables) and applies all pending SQL files under the directory referenced by `VIDFRIENDS_MIGRATIONS` (defaults to `backend/migrations`).

Useful subcommands:

- `go run ./cmd/vidfriends migrate status` – check which migrations have run.
- `go run ./cmd/vidfriends migrate down 1` – roll back the most recent migration. Use carefully; this impacts your database state.

If you are using Docker Compose, make sure the stack is up so the backend CLI can reach the Postgres container:

```bash
cd deploy
docker compose up db -d
```

## Seeding development data

Populate a local database with sample content using the seed script in `backend/seeds/dev_seed.sql`. This script inserts a few demo users, establishes friend relationships, and creates representative video shares so the UI has data out of the box.

Run the seed file with `psql`, substituting your connection URL:

```bash
psql "$VIDFRIENDS_DATABASE_URL" -f backend/seeds/dev_seed.sql
```

The inserted users share the password `password` (hashed with bcrypt). Feel free to customize or extend the script with your own data; rerunning it is safe thanks to the `ON CONFLICT` guards.

## Running the automated tests

Run all backend Go tests from the repository root:

```bash
cd backend
go test ./...
```

For the frontend, install dependencies (if you have not already) and execute the Vite/Jest test runner:

```bash
cd frontend
pnpm install
pnpm test
```

Optionally lint the codebases before submitting changes:

```bash
./scripts/lint.sh
```

Running both backend and frontend test suites before opening a pull request ensures CI passes cleanly and that your environment mirrors the continuous integration pipeline.
