# VidFriends TODO List

A living checklist of follow-up work and enhancements to guide ongoing development.

## Infrastructure & Tooling
- [ ] Review and document the Docker Compose services in `deploy/` to clarify how the backend, frontend, and database interact.
- [ ] Automate CI workflows to run Go tests, frontend unit tests, and linting on every pull request.
- [ ] Set up container build automation that publishes multi-architecture images (x86_64 and arm64).

## Backend
- [ ] Flesh out the Go project structure under `backend/` with handlers, data models, and migrations.
- [ ] Implement authentication endpoints, including session management and token refresh.
- [ ] Add friend management APIs (invite, accept, block) with appropriate database migrations.
- [ ] Integrate yt-dlp metadata lookups for shared video links and cache results.
- [ ] Write comprehensive Go tests for handlers, repositories, and domain logic.

## Frontend
- [ ] Scaffold the React + TypeScript application under `frontend/` with routing and global state management.
- [ ] Implement authentication flows (signup, login, password reset) against the backend APIs.
- [ ] Build friend list management UI with optimistic updates and error handling.
- [ ] Design the shared video feed with metadata display, reactions, and filtering.
- [ ] Add frontend unit and integration tests covering critical user journeys.

## Documentation
- [ ] Expand `docs/STARTUP.md` with troubleshooting tips for common setup issues.
- [ ] Document coding standards and linting/prettier configurations for both backend and frontend.
- [ ] Create contributor guides for running migrations, seeding data, and executing the full test suite locally.
