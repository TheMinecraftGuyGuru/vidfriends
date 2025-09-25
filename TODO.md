# VidFriends TODO Backlog

A curated list of follow-up work to bring the platform from the current prototype
state to the feature set described in the documentation.

## Infrastructure & Tooling
- [x] Add a Makefile or task runner that wraps common workflows (tests, lint,
      database migrations) for both services.
- [x] Configure CI to run `go test ./...`, frontend unit tests, and linting so
      regressions are caught automatically.
- [ ] Provide container images or Docker Compose overrides that bundle the
      `yt-dlp` binary and seed data for local onboarding.

## Backend
- [ ] Implement PostgreSQL-backed repositories for users, friend requests, and
      video shares, then wire them into the HTTP handlers.
- [ ] Replace the in-memory session manager with a persistence-aware solution or
      make it pluggable so access/refresh tokens survive process restarts.
- [ ] Flesh out the migration CLI (`go run ./cmd/vidfriends migrate ...`) so it
      actually runs migrations using the configured database URL.
- [ ] Implement the `/api/v1/videos/feed` endpoint backed by repository queries
      that respect a viewer's friendships.
- [ ] Add password reset endpoints to match the frontend's expectations or
      adjust the client to avoid broken calls.
- [ ] Expand handler and repository test coverage with database integration
      tests (using a test container or transactional rollback strategy).
- [ ] Introduce structured logging and request-scoped context values so errors
      surface actionable metadata.

## Frontend
- [ ] Replace the hard-coded mock data in `AppStateProvider` with real API calls
      for friends, feed entries, and invitations. Provide a mock layer that can
      be toggled for offline development.
- [ ] Align friend invitation mutation URLs with the backend (`/api/v1/friends/respond`)
      and add optimistic update rollbacks when the API rejects changes.
- [ ] Persist and refresh authentication tokens by calling the backend
      `/api/v1/auth/refresh` endpoint before the access token expires.
- [ ] Implement a share-video form that calls the backend and handles metadata
      lookup latency states.
- [ ] Add Vitest/RTL coverage for the dashboard widgets, auth forms, and state
      reducer logic.

## Documentation
- [ ] Update the README and startup guides to reflect the current implementation
      status (e.g. mock data, incomplete endpoints) so new contributors know what
      to expect.
- [ ] Add an API reference that documents available endpoints, request/response
      bodies, and authentication requirements.
- [ ] Document environment variable defaults for both services in a single
      location to reduce duplication between guides.
