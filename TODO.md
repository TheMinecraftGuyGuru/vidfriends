# VidFriends TODO Backlog

A curated list of follow-up work to bring the platform from the current prototype
state to the feature set described in the documentation.

## Infrastructure & Tooling
- [x] Add a Makefile or task runner that wraps common workflows (tests, lint,
      database migrations) for both services.
- [x] Configure CI to run `go test ./...`, frontend unit tests, and linting so
      regressions are caught automatically.
- [x] Provide container images or Docker Compose overrides that bundle the
      `yt-dlp` binary and seed data for local onboarding.
- [x] Replace the placeholder backend and frontend Dockerfiles with builds that
      produce runnable images for the real services.
- [x] Ensure local configuration templates match the documented repo layout
      (e.g. add or document the missing `configs/` directory references).

## Backend
- [x] Implement PostgreSQL-backed repositories for users, friend requests, and
      video shares, then wire them into the HTTP handlers.
- [x] Replace the in-memory session manager with a persistence-aware solution or
      make it pluggable so access/refresh tokens survive process restarts.
- [x] Flesh out the migration CLI (`go run ./cmd/vidfriends migrate ...`) so it
      actually runs migrations using the configured database URL.
- [x] Implement the `/api/v1/videos/feed` endpoint backed by repository queries
      that respect a viewer's friendships.
- [x] Instantiate concrete service dependencies in the HTTP server so handlers
      no longer return "service unavailable" for every request.
- [x] Add password reset endpoints to match the frontend's expectations or
      adjust the client to avoid broken calls.
- [x] Expand handler and repository test coverage with database integration
      tests (using a test container or transactional rollback strategy).
- [x] Introduce structured logging and request-scoped context values so errors
      surface actionable metadata.
- [x] Update the video ingestion pipeline to download and persist assets when
      requested instead of invoking `yt-dlp` with `--skip-download`.
- [x] Extend configuration loading to honor documented object storage (S3/MinIO)
      environment variables.

## Frontend
- [x] Replace the hard-coded mock data in `AppStateProvider` with real API calls
      for friends, feed entries, and invitations. Provide a mock layer that can
      be toggled for offline development.
- [x] Align friend invitation mutation URLs with the backend (`/api/v1/friends/respond`)
      and add optimistic update rollbacks when the API rejects changes.
- [x] Persist and refresh authentication tokens by calling the backend
      `/api/v1/auth/refresh` endpoint before the access token expires.
- [x] Implement a share-video form that calls the backend and handles metadata
      lookup latency states.
- [x] Add Vitest/RTL coverage for the dashboard widgets, auth forms, and state
      reducer logic.

## Documentation
- [x] Update the README and startup guides to reflect the current implementation
      status (e.g. mock data, incomplete endpoints) so new contributors know what
      to expect.
- [x] Add an API reference that documents available endpoints, request/response
      bodies, and authentication requirements.
- [x] Document environment variable defaults for both services in a single
      location to reduce duplication between guides.
- [x] Remove or revise claims about integrated auth/friendship/video features
      until the backend endpoints and frontend wiring exist.
- [x] Clarify setup guides so commands like `go run ./cmd/vidfriends migrate`
      explain their current limitations or are updated once the CLI works.
