# Docker Compose stack

This directory contains the Docker Compose configuration that stitches the
VidFriends backend, frontend, and PostgreSQL database together for local
development. Use it whenever you want to exercise the entire application
without installing the toolchain directly on your workstation.

## File layout

| File | Purpose |
| ---- | ------- |
| `docker-compose.yml` | Defines the multi-service stack. |
| `.env.example` | Template for Compose-level environment variables shared between services. Copy it to `.env` and edit the secrets before running the stack. |

> **Note:** Service-specific `.env` files still live alongside the backend and
> frontend projects. The Compose file mounts them automatically so you can keep
> configuration localized to each codebase.

## Services

The Compose file launches four containers. The diagram below summarizes the
relationships and exposed ports.

```
┌─────────────┐        ┌──────────────┐        ┌──────────────┐
│   frontend  │◀──────▶│   backend    │◀──────▶│     db       │
│  (Vite dev) │ 5173   │  (Go API)    │ 8080   │ PostgreSQL   │ 5432
└─────────────┘        └──────────────┘        └──────────────┘
        ▲                        ▲
        │                        │
        │                        │
        └──────────────┬─────────┘
                       │
                ┌──────┴──────┐
                │   yt-dlp    │
                │ helper img  │
                └─────────────┘
```

### `db`
- **Image**: `postgres:15` (overridable through `POSTGRES_IMAGE`).
- **Data**: Persists to the `db-data` volume so database state survives
  restarts.
- **Env vars**: `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` are
  sourced from `deploy/.env`.
- **Networking**: Exposes port `5432` internally; only the backend container
  should talk to it. If you need direct access, publish `5432:5432` locally.

### `backend`
- **Build context**: `../backend`.
- **Command**: Runs the Go API server (with live reload via `air` if available).
- **Env vars**: Reads `DATABASE_URL`, `SESSION_SECRET`, and `YT_DLP_PATH` from
  `deploy/.env`, plus any backend-specific settings from `backend/.env`.
- **Dependencies**: Waits for `db` to accept connections before starting. It
  performs database migrations automatically on boot.
- **Ports**: Publishes `8080` on the host to serve REST endpoints and websocket
  upgrades.

### `frontend`
- **Build context**: `../frontend`.
- **Command**: Runs the Vite development server so you get hot module reloads
  during UI work.
- **Env vars**: Loads `frontend/.env` to discover the API origin
  (usually `http://localhost:8080`).
- **Ports**: Publishes `5173` to the host. The dev server proxies `/api`
  requests to the backend container.
- **Dependencies**: Depends on the backend so routes that call the API behave as
  expected.

### `yt-dlp`
- **Image**: Lightweight container that only ships the `yt-dlp` binary.
- **Purpose**: Supplies the backend with a pinned version of `yt-dlp` for video
  metadata lookups. The binary is shared over a volume mounted at
  `/usr/local/bin/yt-dlp` in the backend container.
- **Customization**: You can disable this service if you have `yt-dlp` installed
  on the host by updating `YT_DLP_PATH`.

## Usage

```bash
cd deploy
cp .env.example .env   # fill in secrets before the first run
docker compose up --build
```

When the stack is running:

- Visit `http://localhost:5173` for the frontend.
- Hit the API at `http://localhost:8080`.
- Connect to PostgreSQL with `psql` using the credentials defined in `.env`.

Stop services with `docker compose down`. Add `-v` to wipe the database volume.

## Troubleshooting

| Symptom | Likely cause | Fix |
| ------- | ------------ | --- |
| Backend fails with `connection refused` | Database is still starting up | Wait a few seconds or rerun `docker compose up`.
| Frontend can't reach the API | `VITE_API_BASE_URL` misconfigured or backend is down | Update `frontend/.env` or ensure the backend container is healthy.
| Missing `yt-dlp` errors | Helper service disabled or binary not on PATH | Re-enable the `yt-dlp` service or set `YT_DLP_PATH` to a valid binary path. |
