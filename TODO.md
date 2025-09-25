# VidFriends TODO List

A living checklist of follow-up work and enhancements to guide ongoing development.

## Infrastructure & Tooling
- [x] Review and document the Docker Compose services in `deploy/` to clarify how the backend, frontend, and database interact.
- [x] Automate CI workflows to run Go tests, frontend unit tests, and linting on every pull request.
- [x] Set up container build automation that publishes multi-architecture images (x86_64 and arm64).

## Backend
- [x] Flesh out the Go project structure under `backend/` with handlers, data models, and migrations.
- [x] Implement authentication endpoints, including session management and token refresh.
- [x] Add friend management APIs (invite, accept, block) with appropriate database migrations.
- [x] Integrate yt-dlp metadata lookups for shared video links and cache results.
- [x] Write comprehensive Go tests for handlers, repositories, and domain logic.

## Frontend
- [x] Scaffold the React + TypeScript application under `frontend/` with routing and global state management.
- [x] Implement authentication flows (signup, login, password reset) against the backend APIs.
 - [x] Build friend list management UI with optimistic updates and error handling.
- [x] Design the shared video feed with metadata display, reactions, and filtering.
- [x] Add frontend unit and integration tests covering critical user journeys.

## Documentation
- [ ] Expand `docs/STARTUP.md` with troubleshooting tips for common setup issues.
- [ ] Document coding standards and linting/prettier configurations for both backend and frontend.
- [ ] Create contributor guides for running migrations, seeding data, and executing the full test suite locally.
