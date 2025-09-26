# VidFriends

VidFriends is a full-stack video sharing platform with a Go backend, PostgreSQL persistence, and a React + TypeScript frontend.
The project is currently in a **prototype** state: many endpoints are scaffolds, the frontend relies on mock data for most
screens, and key flows such as video ingestion and friendship management still need to be wired up end to end. Use the
documentation in `docs/`—especially the project status notes below—before assuming functionality is complete.

## Getting started

Start with the [First Steps checklist](docs/FIRST_STEPS.md) to orient yourself, then use the
[Startup Guide](docs/STARTUP.md) for detailed instructions on installing prerequisites, configuring environment variables, and
running the stack locally or with Docker Compose. These guides now call out the areas that still depend on mock services or
stubbed commands so you can plan around incomplete functionality. You can also run `./scripts/prepare_workstation.sh` to
validate tool versions and create `.env` files from their templates.

Developers who need a quick reference for configuration should read
[`docs/ENVIRONMENT_VARIABLES.md`](docs/ENVIRONMENT_VARIABLES.md), while API consumers can consult
[`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) for the current surface area and limitations.

## Project structure

- `backend/` – Go service, pgx-powered data layer, migrations, and HTTP handlers.
- `frontend/` – React app that authenticates users, manages friends, and renders the feed UI.
- `migrations/` – SQL migrations executed automatically on service startup.
- `deploy/` – Docker Compose stack and deployment manifests for containerized development.

Refer to the individual directories for component-specific documentation.

## Project status

| Area | Status | Notes |
| ---- | ------ | ----- |
| Authentication & sessions | ⏳ In progress | Backend routes exist but rely on placeholder session storage; expect requests to fail without manual stubbing. |
| Friends & invitations | ⏳ In progress | HTTP handlers are partially implemented and the frontend still serves mock friend data. |
| Video feed | ⏳ In progress | `/api/v1/videos/feed` is scaffolded but not yet backed by the database; the UI renders mock feed entries. |
| Video ingestion | ❌ Not ready | `yt-dlp` integration is not wired up and downloads are skipped. |
| Documentation | ✅ Updated | README, startup guides, and API/environment references reflect the current limitations. |

## Current focus

- Replace frontend mock providers with real API integrations as endpoints come online.
- Finish repository implementations for users, friendships, and video shares so handlers can respond with real data.
- Harden the migration CLI and clarify its limitations until automatic migrations are functional.
- Expand automated testing (Go unit/integration, frontend Vitest) to lock in future work.

## Contributing

1. Fork the repository and create a feature branch.
2. Follow the Startup Guide to spin up the stack and run `go test ./...` and `pnpm test` before committing changes.
3. Open a pull request with a clear description of your changes and testing steps.

Please open an issue if you encounter bugs or have feature requests.
