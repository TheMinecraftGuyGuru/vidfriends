# Environment Variables

This document centralizes the environment variables used by the VidFriends prototype. Values can be provided via `.env` files
(for local development and Docker Compose) or directly in your shell. Defaults are chosen for local development; production
setups should override secrets and endpoints accordingly.

## Backend (`backend/.env`)

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `VIDFRIENDS_PORT` | `8080` | HTTP port the Go API listens on. |
| `VIDFRIENDS_DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/vidfriends?sslmode=disable` | PostgreSQL connection string used by the backend. |
| `VIDFRIENDS_MIGRATIONS` | `migrations` | Path to the directory containing SQL migrations. Relative paths are resolved against the working directory. |
| `VIDFRIENDS_LOG_LEVEL` | `info` | Minimum log level (`debug`, `info`, `warn`, `error`). |
| `VIDFRIENDS_YTDLP_PATH` | `yt-dlp` | Path to the `yt-dlp` binary for metadata lookups. When missing, video creation fails with a 5xx error. |
| `VIDFRIENDS_YTDLP_TIMEOUT` | `30s` | Timeout applied to `yt-dlp` metadata lookups. |
| `VIDFRIENDS_METADATA_CACHE_TTL` | `15m` | Duration that successful metadata lookups are cached in-memory. |
| `VIDFRIENDS_S3_ENDPOINT` | `http://localhost:9000` | MinIO/S3 endpoint used for future asset storage. Not yet fully wired up. |
| `VIDFRIENDS_S3_BUCKET` | `vidfriends` | Default bucket for storing processed video assets. |
| `VIDFRIENDS_S3_REGION` | `us-east-1` | Region passed to the S3 client. |
| `VIDFRIENDS_S3_PUBLIC_BASE_URL` | `http://localhost:9000/vidfriends` | Public URL base for serving stored assets. |
| `SESSION_SECRET` | _none_ | Secret used to sign session cookies. Generate a random 32+ byte string (e.g. `openssl rand -base64 32`). |

## Frontend (`frontend/.env.local`)

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `VITE_API_BASE_URL` | `http://localhost:8080` | Base URL for API requests. Update when running the backend on another host or port. |
| `VITE_USE_MOCKS` | `false` | Enables the mock data layer provided in `configs/frontend.env.example`. Set to `true` while backend endpoints are unfinished. |
| `VITE_USE_MOCK_DATA` | _unset_ | Some parts of the code check this legacy flag. Set it to `true` alongside `VITE_USE_MOCKS` until the variable names are unified. |

> **Tip:** Vite only exposes variables prefixed with `VITE_`. Restart the dev server after changing `.env.local` values.

## Docker Compose (`deploy/.env`)

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `DATABASE_URL` | `postgres://postgres:postgres@db:5432/vidfriends?sslmode=disable` | Database URL injected into the backend container. |
| `SESSION_SECRET` | _none_ | Secret shared with the backend container for session signing. |
| `YT_DLP_PATH` | `/usr/local/bin/yt-dlp` | Path inside the backend container where the helper image installs `yt-dlp`. |
| `BACKEND_PORT` | `8080` | Host port that proxies to the backend container. |
| `FRONTEND_PORT` | `5173` | Host port that proxies to the frontend container. |
| `MINIO_ACCESS_KEY` | `minioadmin` | Access key for the MinIO container. |
| `MINIO_SECRET_KEY` | `minioadmin` | Secret key for the MinIO container. |
| `MINIO_BUCKET` | `vidfriends` | Name of the bucket created by the provisioning job. |

## Usage notes

- Copy the example files from `configs/` (`backend.env.example`, `frontend.env.example`, `deploy.env.example`) to their respective
directories and customize them before running the stack.
- Secrets such as `SESSION_SECRET`, database passwords, and MinIO credentials should never be committed to source control.
- When running tests locally, export the same variables used in development so the services read consistent configuration.
- As new features land, update this document and the example `.env` files to keep the configuration story aligned.
