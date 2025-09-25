# VidFriends Startup Guide

This guide describes how to bootstrap the VidFriends development environment in two different ways:

1. **Local toolchain** – run each service (Go API, PostgreSQL, React frontend) directly on your machine.
2. **Docker Compose** – start the entire stack with containers and minimal host dependencies.

Use whichever method best matches your workflow. The Docker Compose path is usually fastest for first-time contributors, while the local toolchain workflow makes iterative backend/frontend development easier.

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
| [yt-dlp](https://github.com/yt-dlp/yt-dlp#installation) | latest | CLI utility used by the backend for video metadata lookups and downloading shared videos for storage. |

Optional but recommended:

- [golangci-lint](https://golangci-lint.run/welcome/install/) to run the Go linters locally.
- [Air](https://github.com/cosmtrek/air) or a similar hot-reload tool for the Go service.
- [mkcert](https://github.com/FiloSottile/mkcert) if you want to test HTTPS locally.

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

Some of these directories are generated during development or may be moved as the project evolves, but the startup steps below assume this default layout.

---

## 3. Configure environment variables

Both the backend service and the Docker Compose deployment read configuration from environment variables. The repository includes example files that you can copy and customize:

```bash
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
cp deploy/.env.example deploy/.env
```

Update the copied files with values that match your environment. At a minimum you need to provide:

- `DATABASE_URL` – PostgreSQL connection string used by the Go service (defaults to
  `postgres://postgres:postgres@localhost:5432/vidfriends?sslmode=disable`).
- `SESSION_SECRET` – random 32+ byte string for signing session cookies. Generate
  a new value with `openssl rand -base64 32`.
- `YT_DLP_PATH` – optional path to the `yt-dlp` binary if it is not on `$PATH`.
- `VITE_API_BASE_URL` – API origin for the frontend (`http://localhost:8080` in
  local development).
- `VITE_USE_MOCKS` – toggle the frontend's mock service layer (`false` by
  default).
- `POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` – credentials the
  Docker Compose workflow applies to the PostgreSQL container.
- `BACKEND_PORT` / `FRONTEND_PORT` – override ports exposed by Docker Compose if
  `8080` or `5173` conflict with other services on your machine.

> **Tip:** Generate secure secrets with `openssl rand -base64 32`.

---

## 4. Local development workflow

Follow these steps to run the stack directly on your machine.

### 4.1 Start PostgreSQL

Create a database role and schema for VidFriends. The example below uses the default PostgreSQL superuser:

```bash
createdb vidfriends
psql -d vidfriends -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
```

Update `DATABASE_URL` in `backend/.env` to point at this database. For example:

```ini
DATABASE_URL=postgres://localhost:5432/vidfriends?sslmode=disable
```

### 4.2 Run database migrations

The Go service automatically runs migrations on startup, but you can run them manually for faster feedback:

```bash
cd backend
go run ./cmd/vidfriends migrate up
```

This command uses the `migrations/` directory to apply schema changes. To roll back:

```bash
go run ./cmd/vidfriends migrate down 1
```

### 4.3 Launch the Go API

With the environment configured, start the backend in development mode:

```bash
cd backend
source .env
go run ./cmd/vidfriends serve
```

The service listens on port `8080` by default and exposes REST endpoints for authentication, friend management, video sharing, and feed retrieval. Logs indicate when migrations and dependency checks succeed.

### 4.4 Start the React frontend

In a separate terminal:

```bash
cd frontend
pnpm install   # or npm install
touch .env.local  # ensure the file exists, then edit as needed
pnpm dev       # or npm run dev
```

By default the development server runs on `http://localhost:5173`. It proxies API requests to the Go backend using `VITE_API_BASE_URL`.

### 4.5 Verify the stack

1. Visit `http://localhost:5173` and sign up for a new account.
2. Add a friend using their username or email.
3. Share a video link to confirm yt-dlp metadata retrieval and download-to-storage workflow.
4. Check the "Feed" tab to ensure shared videos appear.

The backend logs should show API traffic and share fan-out operations, while the database should record new users, sessions, and shares.

---

## 5. Docker Compose workflow

The repository ships with `deploy/docker-compose.yml` to orchestrate the services. Environment values are pulled from `deploy/.env` and the `.env` files in the backend/frontend directories.

### 5.1 Build and start containers

```bash
cd deploy
cp .env.example .env  # if you haven't already
# Edit .env to set secrets like DATABASE_URL, SESSION_SECRET, etc.
docker compose up --build
```

This command launches the following containers:

- `db` – PostgreSQL with initialized databases and volumes for persistence.
- `backend` – Go API service with hot-reload support (via `air`) mounting your source tree.
- `frontend` – React dev server served via Vite.
- `yt-dlp` – optional helper image that caches the `yt-dlp` binary.

Wait until the logs show the API listening and migrations completing. The frontend is accessible at `http://localhost:5173` and the API at `http://localhost:8080`.

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

## 6. Testing and linting

Run the automated checks before submitting a pull request.

```bash
cd backend
go test ./...
# Optional linters if golangci-lint is installed
golangci-lint run

cd ../frontend
pnpm test
pnpm lint
```

These commands ensure the Go API, React frontend, and TypeScript sources compile and pass their respective test suites.

---

## 7. Troubleshooting

The matrix below captures the issues we see most often and how to recover from them. Many fixes involve re-reading the `.env`
files for typos or missing values, so start there before digging into deeper debugging.

| Symptom | Likely cause | Suggested fix |
| ------- | ------------ | ------------- |
| API fails to start with `pq: password authentication failed` | Incorrect database credentials | Double-check `DATABASE_URL` and PostgreSQL role configuration. |
| Go API exits with `dial tcp 127.0.0.1:5432: connect: connection refused` | PostgreSQL server is not running or listening on a different port | Start PostgreSQL (`brew services start postgresql` / `systemctl start postgresql`) and confirm the port matches `DATABASE_URL`. |
| `go run ./cmd/vidfriends serve` reloads repeatedly in Docker | The source tree is mounted with root-owned files, causing `air` to fail watching | Run `sudo chown -R $(id -u):$(id -g) .` inside the repo or adjust the bind mount user in `docker-compose.yml`. |
| `yt-dlp` errors when fetching metadata or downloading files | Missing binary, insufficient disk space, or rate-limiting | Install `yt-dlp` locally, ensure the storage volume has capacity, or set `YT_DLP_PATH` to a bundled binary. Wait/retry if the provider is rate-limiting. |
| `pnpm install` fails with "unsupported engine" | Your Node.js runtime is older than the minimum version | Upgrade Node.js to the current LTS release (18+) and re-run `pnpm install`. |
| CORS errors in the browser console | Frontend origin not in backend `CORS_ALLOWED_ORIGINS` | Update the backend `.env` to include the frontend dev origin. |
| Session cookie missing in requests | Browser blocked third-party cookies or secure flag mismatch | Use the same domain/port in development and ensure `SESSION_SECURE=false` for HTTP. |
| Frontend shows blank screen with Vite `ERR_NETWORK` errors | `VITE_API_BASE_URL` points to the wrong host/port | Confirm the API is reachable from the browser and update `frontend/.env.local` (for Docker ensure you use the host machine IP). |
| `docker compose up` exits immediately | Ports already in use | Stop conflicting services or change the exposed ports in the compose file. |

### 7.1 Logs and diagnostics

- **Backend logs:** Run `go run ./cmd/vidfriends serve --log-level debug` (or `docker compose logs -f backend`) to view detailed
  request traces and configuration warnings.
- **Database connectivity:** `psql $DATABASE_URL -c '\dt'` verifies the service can reach PostgreSQL and that migrations have
  been applied.
- **Frontend diagnostics:** Open the browser dev tools console and network tab to confirm the API responses and CORS headers
  match your expectations.
- **Docker environment:** `docker compose ps` shows container status, and `docker compose exec backend env` can be used to
  inspect environment variables inside the running service.

### 7.2 When all else fails

- Rebuild containers without cache: `docker compose build --no-cache` followed by `docker compose up`.
- Remove dangling volumes when schema drift causes migration failures: `docker compose down -v` (this wipes the local
  database).
- Delete and recreate `.env` files from the provided `.example` templates to ensure no stale keys remain.
- Ask for help in the project chat with the exact command, full log output, and details about your OS, CPU, and tooling
  versions.

---

## 8. Next steps

- Review the API documentation in `backend/docs/openapi.yaml` (generated via `swag init`).
- Update the sample `.env.example` files when configuration options change.
- Consider adding `make` targets for the most common tasks (e.g., `make dev`, `make test`).

If you run into issues not covered here, open a GitHub issue with details about your environment, the commands you ran, and any error output.

