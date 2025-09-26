# VidFriends Startup Guide

This guide describes how to bootstrap the VidFriends development environment in two different ways. The stack is still under
active development—several flows rely on mock data, migrations are evolving, and yt-dlp downloads are not yet persisted—so the
steps below call out known limitations to help you plan your testing strategy.

1. **Local toolchain** – run each service (Go API, PostgreSQL, React frontend) directly on your machine.
2. **Docker Compose** – start the entire stack with containers and minimal host dependencies.

Use whichever method best matches your workflow. The Docker Compose path is usually fastest for first-time contributors, while
the local toolchain workflow makes iterative backend/frontend development easier. Keep `VITE_USE_MOCKS=true` if you want
components that do not yet have APIs to render placeholder content, but note that the global app state now always talks to the
real backend.

---

## 1. Prerequisites

### Local toolchain prerequisites

Install the following software before you begin:

| Tool | Minimum Version | Purpose |
| ---- | --------------- | ------- |
| [Go](https://go.dev/dl/) | 1.21 | Builds and runs the backend service. |
| [Node.js](https://nodejs.org/en/download/) | 18 LTS | Runs the React development server and tooling. |
| [pnpm](https://pnpm.io/installation) or [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) | latest | Installs frontend dependencies. |
| [PostgreSQL](https://www.postgresql.org/download/) | 14+ | Local development database. |
| [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) | latest | CLI utility used by the backend for video metadata lookups. Downloads are still skipped in development builds. |

Optional but recommended:

- [golangci-lint](https://golangci-lint.run/welcome/install/) to run the Go linters locally.
- [Air](https://github.com/cosmtrek/air) or a similar hot-reload tool for the Go service.
- [mkcert](https://github.com/FiloSottile/mkcert) if you want to test HTTPS locally.
- [`httpie`](https://httpie.io/) or [`curl`](https://curl.se/) for exercising API endpoints without the frontend mocks.

### Docker-based prerequisites

If you prefer running everything inside containers, install:

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) **or** the Docker Engine & Docker Compose plugin on Linux.
- [`make`](https://www.gnu.org/software/make/manual/make.html) (optional) to use the convenience targets described below.

---

## 2. Repository layout

The repository is organized as follows:

```
backend/         # Go API source (cmd/, internal/, pkg/)
frontend/        # React + TypeScript SPA
migrations/      # SQL migration files applied at startup
deploy/          # Docker Compose, container build scripts, and k8s manifests
configs/         # Sample configuration and environment files
```

Some of these directories are generated during development or may be moved as the project evolves, but the startup steps below
assume this default layout.

---

## 3. Configure environment variables

Both the backend service and the Docker Compose deployment read configuration from environment variables. The repository includes
example files that you can copy and customize. See [`docs/ENVIRONMENT_VARIABLES.md`](ENVIRONMENT_VARIABLES.md) for a full catalog
of supported options and defaults:

```bash
cp configs/backend.env.example backend/.env
cp configs/frontend.env.example frontend/.env.local
cp configs/deploy.env.example deploy/.env
```

Update the copied files with values that match your environment. At a minimum you need to provide:

- `DATABASE_URL` – PostgreSQL connection string used by the Go service (defaults to
  `postgres://postgres:postgres@localhost:5432/vidfriends?sslmode=disable`).
- `SESSION_SECRET` – random 32+ byte string for signing session cookies. Generate
  a new value with `openssl rand -base64 32`.
- `YT_DLP_PATH` – optional path to the `yt-dlp` binary if it is not on `$PATH`. Metadata lookups fall back to mock data if the
  binary is missing.
- `VITE_API_BASE_URL` – API origin for the frontend (`http://localhost:8080` in
  local development). The AppStateProvider always uses this endpoint for authentication, friends, and feed data.
- `VITE_USE_MOCKS` – toggle the remaining mock service layer (`false` by default). Set this to `true` when you need placeholder content for features without finished APIs.
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` – credentials the
  Docker Compose workflow applies to the PostgreSQL container.
- `BACKEND_PORT` / `FRONTEND_PORT` – override ports exposed by Docker Compose if
  `8080` or `5173` conflict with other services on your machine.

> **Tip:** Generate secure secrets with `openssl rand -base64 32`.

---

## 4. Local development workflow

Follow these steps to run the stack directly on your machine.

### 4.1 Start PostgreSQL

Create a database role and schema for VidFriends. The example below uses the default PostgreSQL superuser. Until migrations are
finalized you may need to drop and recreate the database between iterations:

```bash
createdb vidfriends
psql -d vidfriends -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
```

Update `DATABASE_URL` in `backend/.env` to point at this database. For example:

```ini
DATABASE_URL=postgres://localhost:5432/vidfriends?sslmode=disable
```

### 4.2 Run database migrations

The Go service automatically runs migrations on startup, but you can run them manually for faster feedback. Downward migrations
are not yet supported, so re-apply by recreating the database if something goes wrong:

```bash
cd backend
go run ./cmd/vidfriends migrate up
```

This command uses the `migrations/` directory to apply schema changes in sorted order. Use `go run ./cmd/vidfriends migrate status`
to list applied migrations.

### 4.3 Launch the Go API

With the environment configured, start the backend in development mode:

```bash
cd backend
source .env
go run ./cmd/vidfriends serve
```

The service listens on port `8080` by default and exposes REST endpoints for authentication, friend management, video sharing, and
feed retrieval. Logs indicate when migrations and dependency checks succeed. Expect non-critical routes (like password reset) to
respond with placeholder behavior until their integrations are completed.

### 4.4 Start the React frontend

In a separate terminal:

```bash
cd frontend
pnpm install   # or npm install
touch .env.local  # ensure the file exists, then edit as needed
pnpm dev       # or npm run dev
```

By default the development server runs on `http://localhost:5173`. It proxies API requests to the Go backend using
`VITE_API_BASE_URL`. Ensure the backend is running; otherwise authentication, friends, and feed views will surface empty
states. You can still enable `VITE_USE_MOCKS=true` to stub out components that lack endpoints.

### 4.5 Verify the stack

1. Visit `http://localhost:5173` and sign up for a new account. If anything fails, file an issue rather than switching back to
   the legacy mock data.
2. Add a friend using their username or email and respond to invitations. The UI now mirrors the backend state instead of local
   fixtures.
3. Share a video link to confirm yt-dlp metadata retrieval. Downloads are skipped for now, but metadata lookups should succeed
   when `yt-dlp` is available.
4. Check the "Feed" tab to ensure shared videos appear for your account.

The backend logs should show API traffic, even if some handlers return TODO responses. Track unexpected failures in the roadmap so
they can be prioritized.

---

## 5. Docker Compose workflow

The repository ships with `deploy/docker-compose.yml` to orchestrate the services. Environment values are pulled from `deploy/.env`
and the `.env` files in the backend/frontend directories.

### 5.1 Build and start containers

```bash
cd deploy
cp .env.example .env  # if you haven't already
# Edit .env to set secrets like DATABASE_URL, SESSION_SECRET, and MinIO credentials.
docker compose up --build --remove-orphans
```

This command launches the following containers (some provide scaffolding only and are not yet wired for production workloads):

- `db` – PostgreSQL with initialized databases and volumes for persistence.
- `backend` – Go API service running `go run ./cmd/vidfriends serve` with the repo mounted as a volume. Uses mocks for yt-dlp
  downloads until the ingestion pipeline lands.
- `frontend` – React dev server served via Vite. Enable mocks via `.env.local` to avoid errors while backend features remain in
  flight.
- `minio` – S3-compatible object storage that holds processed video assets. Ingestion is stubbed, so expect empty buckets.
- `yt-dlp` – helper image that copies the `yt-dlp` binary onto a shared volume.
- `createbuckets` – one-time MinIO client job that provisions the public bucket defined in `.env`.

Wait until the logs show the API listening, the bucket provisioning job exiting successfully, and PostgreSQL reporting healthy. The
frontend is accessible at `http://localhost:5173`, the API at `http://localhost:8080`, and the object storage browser at
`http://localhost:9001`. Because MinIO and yt-dlp are not yet hooked into an ingestion workflow, you can ignore warnings from those
containers during early development.

### 5.2 Running commands inside containers

Use `docker compose exec` for ad-hoc tasks:

```bash
docker compose exec backend go test ./...
docker compose exec backend go run ./cmd/vidfriends migrate up
docker compose exec frontend pnpm test
```

### 5.3 Stopping and cleaning up

```bash
docker compose down           # stop services
docker compose down -v        # stop and remove volumes (wipes the database)
```

---

Following these steps gets you a working development environment with realistic expectations about what is still a work in
progress. As endpoints stabilize you can disable the mock layers and tighten the verification checklist.
