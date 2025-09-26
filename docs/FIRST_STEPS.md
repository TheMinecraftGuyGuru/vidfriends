# First steps for the VidFriends project

This checklist helps a new contributor or maintainer get oriented, prepare a workstation, and confirm that the VidFriends stack
can run locally. Work through the sections sequentially and tick each box as you finish the activity. Items now call out whether
they depend on in-progress features so you can skip or mock them when necessary.

## 1. Understand the architecture

- [ ] Read the top-level [repository overview](../README.md) to learn what lives
      in `backend/`, `frontend/`, `docs/`, `deploy/`, and `scripts/` along with the current project status.
- [ ] Skim [`backend/README.md`](../backend/README.md) and
      [`frontend/README.md`](../frontend/README.md) to understand how each
      service is started, tested, and how the directories are laid out. Take note of TODO sections that mark unfinished flows.
- [ ] Review [`docs/STARTUP.md`](STARTUP.md) so you know how the backend,
      frontend, database, and tooling are expected to interact. Pay special attention to the mock-mode guidance and migration
      limitations.
- [ ] Familiarize yourself with [`docs/API_REFERENCE.md`](API_REFERENCE.md) to learn which endpoints are ready for use and which
      still return placeholder responses.

## 2. Prepare your development workstation

- [ ] Clone the repository locally and, if you plan to submit pull requests,
      create a personal fork on GitHub.
- [ ] Run `./scripts/prepare_workstation.sh` to validate minimum Go/Node tool
      versions and automatically copy `.env` templates into the backend,
      frontend, and `deploy/` directories. The script will highlight optional
      tools (yt-dlp, Docker, linting) that unlock more of the workflow once
      those features are wired up.
- [ ] Open the generated environment files (`backend/.env`,
      `frontend/.env.local`, and `deploy/.env`) and replace placeholder values
      with secrets and connection strings that match your local setup.
- [ ] Install any optional tooling called out by the script output (e.g.
      `golangci-lint`, `yt-dlp`, or Docker) so future steps run smoothly—even if
      some integrations are still mocked.

## 3. Run the stack locally

- [ ] Start PostgreSQL and create a `vidfriends` database (or adjust
      `DATABASE_URL` in `backend/.env` to point at an existing instance).
- [ ] Apply database migrations with `go run ./cmd/vidfriends migrate up`. Down migrations are not supported yet, so reset the
      database if you need to re-apply.
- [ ] Launch the Go API with `go run ./cmd/vidfriends serve` and confirm it
      listens on the port configured in your `.env` (default `8080`). Capture any
      `TODO` log statements or unimplemented responses in the roadmap.
- [ ] Install frontend dependencies with `pnpm install` (or `npm install`) and
      run `pnpm dev` so the Vite dev server starts on port `5173`.
- [ ] Visit `http://localhost:5173`, enable the mock service layer if the API is
      still stabilizing, register a throwaway account, add a friend, and share a
      video link. Document which flows work end-to-end and which still depend on
      mocks.

## 4. Automate container builds

- [ ] Populate `deploy/` with a Docker Compose workflow (or confirm the existing
      definition still matches the stack) so contributors can bring the services
      up with `docker compose up --build`. Note any containers that currently serve as placeholders.
- [ ] Provision a GitHub-hosted or self-hosted CI runner with Docker Buildx and
      QEMU so multi-architecture images can be built and pushed once the services are production-ready.
- [ ] Update the repository's automation (e.g. GitHub Actions) to build and
      publish backend and frontend images on pushes and pull requests. Skipping this is acceptable while the stack is incomplete,
      but track it in the roadmap.
- [ ] Ensure the build pipeline produces images for both `linux/amd64` and
      `linux/arm64` so developers on Intel and Apple Silicon hardware can run
      the containers when the services are ready.

## 5. Establish team workflows

- [ ] Agree on coding standards, linting rules, and testing expectations for the
      backend and frontend, then document them in the repository.
- [ ] Configure branch protection rules and required status checks for `main`.
- [ ] Groom open issues, prioritize a roadmap, and assign work to collaborators
      so everyone knows what to focus on next. Use the roadmap checklist to capture incomplete backend endpoints, frontend mock
      replacements, and documentation debt.

Completing these steps ensures you have a reliable local environment, understand
how the pieces fit together, and have automation plus team processes ready for
collaborative development—even while portions of the product are still under construction.
