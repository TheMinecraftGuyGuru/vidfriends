# Roadmap to a Working VidFriends Release

Use this checklist to prioritize the remaining engineering work required to move from today's prototype to a fully functional
MVP. Items are grouped roughly in the order they should be tackled; update the list as milestones are completed.

## Backend stabilization

- [x] Finalize database migrations (users, sessions, friend requests, video shares) and add seed data for local onboarding.
- [ ] Harden the migration CLI with rollback support or transactional retries so failed runs can be recovered without dropping the database.
- [ ] Wire structured logging and tracing context into all handlers to simplify debugging during integration testing.
- [ ] Implement background jobs to persist video assets to object storage once `yt-dlp` metadata retrieval succeeds.
- [ ] Add rate limiting and input validation guards to authentication and invite endpoints.

## Frontend integration

- [ ] Replace the AppStateProvider mock toggles with real API calls once the backend endpoints are stable.
- [ ] Implement refresh-token handling in the frontend so sessions stay alive without manual reloads.
- [ ] Build optimistic UI flows (and rollbacks) for friend invitations and video shares.
- [ ] Add error boundary and toast messaging for API failures surfaced by the new backend responses.

## Cross-cutting concerns

- [ ] Document end-to-end manual test cases that verify signup, login, friend invites, and sharing flows.
- [ ] Expand automated test coverage: backend integration tests against PostgreSQL, frontend Vitest + React Testing Library suites.
- [ ] Set up CI workflows that run Go tests, frontend tests, linting, and type checks on every pull request.
- [ ] Establish staging Docker images (backend + frontend) published from main to exercise deployment paths.

## Launch readiness

- [ ] Populate production-ready configuration templates (secrets management, TLS, MinIO/S3 credentials).
- [ ] Conduct load testing on the feed and share endpoints to size infrastructure requirements.
- [ ] Prepare onboarding documentation for early adopters, including known limitations and troubleshooting steps.
- [ ] Schedule a bug bash or friend-and-family beta once the above checkpoints are complete.

Track progress weekly and refine the roadmap as new information emerges. Update this file whenever priorities change so the team
always shares the same picture of "done."
