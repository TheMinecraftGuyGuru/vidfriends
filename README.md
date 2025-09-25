# VidFriends

VidFriends is a full-stack video sharing platform with a Go backend, PostgreSQL persistence, and a React + TypeScript frontend. The
backend exposes REST endpoints for authentication, friendship management, and sharing video links with yt-dlp metadata lookups,
while the frontend delivers a responsive, dark-themed single-page app for managing your feed.

## Getting started

Start with the [First Steps checklist](docs/FIRST_STEPS.md) to orient yourself, then use the
[Startup Guide](docs/STARTUP.md) for detailed instructions on installing prerequisites, configuring environment
variables, and running the stack locally or with Docker Compose. You can also run
`./scripts/prepare_workstation.sh` to validate tool versions and create `.env` files from their templates.

## Project structure

- `backend/` – Go service, pgx-powered data layer, migrations, and HTTP handlers.
- `frontend/` – React app that authenticates users, manages friends, and renders the feed UI.
- `migrations/` – SQL migrations executed automatically on service startup.
- `deploy/` – Docker Compose stack and deployment manifests for containerized development.

Refer to the individual directories for component-specific documentation.

## Feature set

- For You Page-style video feed that surfaces shared clips in an immersive, swipeable stream.
- Authentication, friendship management, and sharing flows tightly integrated across the stack.
- Responsive, dark-themed UI optimized for both desktop and mobile devices.

## Roadmap

- Iterate on the For You Page feed experience with richer metadata, reactions, and personalization controls.
- Expand collaboration features, including co-watching sessions and friend recommendations.
- Harden observability, automated testing, and deployment workflows for production readiness.

## Contributing

1. Fork the repository and create a feature branch.
2. Follow the Startup Guide to spin up the stack and run `go test ./...` and `pnpm test` before committing changes.
3. Open a pull request with a clear description of your changes and testing steps.

Please open an issue if you encounter bugs or have feature requests.

