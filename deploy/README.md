# Docker Compose stack

This directory contains the Docker Compose configuration that stitches the
VidFriends backend, frontend, and PostgreSQL database together for local
development. Use it whenever you want to exercise the entire application
without installing the toolchain directly on your workstation.

## File layout

| File | Purpose |
| ---- | ------- |
| `docker-compose.yml` | Defines the multi-service stack. |
| `docker-compose.onboarding.yml` | Optional override that bundles `yt-dlp` and database seed data for demos. |
| `.env.example` | Symlink to `../configs/deploy.env.example`, the template for Compose-level environment variables shared between services. Copy it to `.env` and edit the secrets before running the stack. |
| `docker/` | Dockerfiles used by the CI pipeline to build multi-architecture backend and frontend images. |

> **Note:** Service-specific `.env` files still live alongside the backend and
> frontend projects. The Compose file mounts them automatically so you can keep
> configuration localized to each codebase.

## Services

The Compose file launches five long-lived containers plus a short-lived job
that provisions MinIO buckets. The diagram below summarizes the relationships
and exposed ports.

```
┌─────────────┐        ┌──────────────┐        ┌──────────────┐
│   frontend  │◀──────▶│   backend    │◀──────▶│     db       │
│  (Vite dev) │ 5173   │  (Go API)    │ 8080   │ PostgreSQL   │ 5432
└─────────────┘        └──────▲───────┘        └──────────────┘
        ▲                     │
        │                     │
        │                     │
        │            ┌────────┴──────────┐
        │            │      MinIO        │
        │            │  S3 Storage 9000  │
        │            └────────▲──────────┘
        │                     │
        └──────────────┬──────┘
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
- **Networking**: Exposes port `5432` inside the Compose network; publish
  `${POSTGRES_PORT:-5432}` to access it from the host.

### `backend`
- **Image**: `golang:1.22` running `go run ./cmd/vidfriends serve` against the
  mounted source tree.
- **Env vars**: Reads `DATABASE_URL`, `SESSION_SECRET`, and `YT_DLP_PATH` from
  `deploy/.env`, plus any backend-specific settings from `backend/.env`.
- **Dependencies**: Waits for PostgreSQL and the MinIO bucket job to complete
  before starting. Migrations currently need to be run manually.
- **Ports**: Publishes `${BACKEND_PORT:-8080}` on the host to serve REST
  endpoints and websocket upgrades.

### `frontend`
- **Image**: `node:20` with `pnpm` enabled via Corepack.
- **Command**: Installs dependencies and runs the Vite development server so you
  get hot module reloads during UI work.
- **Env vars**: Loads `frontend/.env` to discover the API origin (usually
  `http://localhost:8080`).
- **Ports**: Publishes `${FRONTEND_PORT:-5173}` to the host. The dev server
  proxies `/api` requests to the backend container.
- **Dependencies**: Depends on the backend so routes that call the API behave as
  expected.

### `yt-dlp`
- **Image**: `ghcr.io/yt-dlp/yt-dlp:2023.11.14` by default (overridden to a
  custom build when the onboarding profile is enabled).
- **Purpose**: Copies the yt-dlp binary onto a shared volume. The backend mounts
  that volume at `/usr/local/bin/yt-dlp` so video metadata lookups and downloads
  use the pinned binary. Under the onboarding override, the helper image also
  publishes SQL migrations and seed data for the database bootstrap job.
- **Customization**: You can disable this service if you have `yt-dlp` installed
  on the host by updating `YT_DLP_PATH`.

### `minio`
- **Image**: `minio/minio:RELEASE.2024-01-13T07-53-03Z`.
- **Purpose**: Provides S3-compatible object storage for downloaded videos and
  thumbnails. The API listens on `${MINIO_API_PORT:-9000}` and the admin console
  is available on `${MINIO_CONSOLE_PORT:-9001}`.
- **Credentials**: Set `MINIO_ROOT_USER` and `MINIO_ROOT_PASSWORD` in
  `deploy/.env` before starting the stack.

### `createbuckets`
- **Image**: `minio/mc:RELEASE.2024-01-11T07-46-16Z`.
- **Purpose**: One-time job that provisions the bucket defined by
  `VIDFRIENDS_S3_BUCKET` and makes it world-readable so videos can be streamed
  directly from the MinIO endpoint.
- **Restart policy**: Runs to completion and exits.

## Configuration reference

The Compose stack is configured entirely through environment variables sourced
from `deploy/.env` and the service-specific `.env` files. The canonical list of
supported settings and their defaults lives in [`deploy/settings.json`](settings.json).
The tables below summarise the values you can override when launching the stack.

### Container images

| Variable | Default | Description |
| --- | --- | --- |
| `POSTGRES_IMAGE` | `postgres:15` | Image used for the PostgreSQL database service. |
| `BACKEND_IMAGE` | `golang:1.22` | Base image that runs the Go backend in Compose. |
| `FRONTEND_IMAGE` | `node:20` | Image that powers the Vite development server. |
| `MINIO_IMAGE` | `minio/minio:RELEASE.2024-01-13T07-53-03Z` | MinIO object storage container image. |
| `MINIO_MC_IMAGE` | `minio/mc:RELEASE.2024-01-11T07-46-16Z` | MinIO client image that provisions buckets. |
| `YTDLP_IMAGE` | `ghcr.io/yt-dlp/yt-dlp:2023.11.14` | Helper image used to publish the `yt-dlp` binary onto a shared volume. |

### Database connectivity

| Variable | Default | Description |
| --- | --- | --- |
| `POSTGRES_USER` | `postgres` | Database superuser created by the `db` service. |
| `POSTGRES_PASSWORD` | `postgres` | Password associated with `POSTGRES_USER`. |
| `POSTGRES_DB` | `vidfriends` | Database created on first launch. |
| `POSTGRES_PORT` | `5432` | Host port that publishes PostgreSQL outside of Docker. |
| `DATABASE_URL` | `postgres://postgres:postgres@db:5432/vidfriends?sslmode=disable` | Connection string the backend uses to reach PostgreSQL inside the Compose network. |

### Backend runtime

| Variable | Default | Description |
| --- | --- | --- |
| `BACKEND_PORT` | `8080` | Host port published for the Go API. |
| `SESSION_SECRET` | `replace-with-secure-random-string` | Secret used to sign and verify session cookies. |
| `VIDFRIENDS_LOG_LEVEL` | `info` | Structured logging level emitted by the service. |
| `VIDFRIENDS_MIGRATIONS` | `migrations` | Directory containing SQL migration files inside the container. |
| `YT_DLP_PATH` | `/usr/local/bin/yt-dlp` | Filesystem path to the `yt-dlp` binary mounted in the backend container. |
| `VIDFRIENDS_YTDLP_TIMEOUT` | `30s` | Maximum time the backend waits for `yt-dlp` metadata responses. |
| `VIDFRIENDS_METADATA_CACHE_TTL` | `15m` | Lifetime of cached video metadata entries. |
| `VIDFRIENDS_S3_ENDPOINT` | `http://minio:9000` | Internal endpoint used to talk to MinIO. |
| `VIDFRIENDS_S3_BUCKET` | `vidfriends` | Bucket name created for storing processed assets. |
| `VIDFRIENDS_S3_REGION` | `us-east-1` | Region identifier reported to S3 clients. |
| `VIDFRIENDS_S3_PUBLIC_BASE_URL` | `http://localhost:9000/vidfriends` | Base URL used to build public links to stored videos. |

### Frontend runtime

| Variable | Default | Description |
| --- | --- | --- |
| `FRONTEND_PORT` | `5173` | Host port used by the Vite dev server. |
| `FRONTEND_API_BASE_URL` | `http://backend:8080` | API origin the frontend proxies requests to while running in Compose. |

### Object storage credentials

| Variable | Default | Description |
| --- | --- | --- |
| `MINIO_ROOT_USER` | `vidfriends` | Administrative access key for MinIO. |
| `MINIO_ROOT_PASSWORD` | `vidfriends-secret` | Administrative secret key for MinIO. |
| `MINIO_API_PORT` | `9000` | Host port exposing the MinIO S3-compatible API. |
| `MINIO_CONSOLE_PORT` | `9001` | Host port exposing the MinIO management console. |

## Usage

```bash
cd deploy
cp .env.example .env   # fill in secrets before the first run
docker compose up --build --remove-orphans
```

Need a turnkey demo database and bundled `yt-dlp` binary? Combine the base file
with the onboarding override:

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.onboarding.yml \
  up --build --remove-orphans
```

The override swaps in an assets helper image that publishes the pinned
`yt-dlp` executable to the shared volume, applies SQL migrations, and seeds the
database before the backend starts accepting requests.

When the stack is running:

- Visit `http://localhost:5173` for the frontend.
- Hit the API at `http://localhost:8080`.
- Browse stored videos at `http://localhost:9000/${VIDFRIENDS_S3_BUCKET}` or use the MinIO console at `http://localhost:9001`.
- Connect to PostgreSQL with `psql` using the credentials defined in `.env`.

Stop services with `docker compose down`. Add `-v` to wipe the database volume.

## Automated image builds

The repository publishes multi-architecture container images (x86_64/amd64 and
arm64) for the backend and frontend whenever changes land on `main`, a release
is published, or the workflow is triggered manually. The GitHub Actions workflow
[`container-build.yml`](../.github/workflows/container-build.yml) uses Docker
Buildx with QEMU emulation to produce and push manifests to the GitHub Container
Registry under `ghcr.io/<org>/<repo>-<service>`.

If you need to test builds locally with the same configuration, install
Docker Buildx and run:

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file deploy/docker/backend.Dockerfile \
  --tag ghcr.io/<your-namespace>/vidfriends-backend:dev \
  .
```

Replace the Dockerfile path and tag for the frontend image as needed.

## Troubleshooting

| Symptom | Likely cause | Fix |
| ------- | ------------ | --- |
| Backend fails with `connection refused` | Database is still starting up | Wait a few seconds or rerun `docker compose up`.
| Frontend can't reach the API | `VITE_API_BASE_URL` misconfigured or backend is down | Update `frontend/.env` or ensure the backend container is healthy.
| Missing `yt-dlp` errors | Helper service disabled or binary not on PATH | Re-enable the `yt-dlp` service or set `YT_DLP_PATH` to a valid binary path. |
| MinIO bucket missing or private | Bucket provisioning job failed or credentials changed | Re-run `docker compose run --rm createbuckets` after updating the MinIO secrets. |
