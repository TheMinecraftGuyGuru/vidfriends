# First steps for the VidFriends project

This checklist helps a new contributor or maintainer get oriented, prepare a
workstation, and confirm that the VidFriends stack can run locally. Work through
the sections sequentially and tick each box as you finish the activity.

## 1. Understand the architecture

- [ ] Read the top-level [repository overview](../README.md) to learn what lives
      in `backend/`, `frontend/`, `docs/`, `deploy/`, and `scripts/`.
- [ ] Skim [`backend/README.md`](../backend/README.md) and
      [`frontend/README.md`](../frontend/README.md) to understand how each
      service is started, tested, and how the directories are laid out.
- [ ] Review [`docs/STARTUP.md`](STARTUP.md) so you know how the backend,
      frontend, database, and tooling are expected to interact. Take note of any
      referenced assets that still need to be implemented (for example, the
      Docker Compose definition or SQL migration files).

## 2. Prepare your development workstation

- [ ] Clone the repository locally and, if you plan to submit pull requests,
      create a personal fork on GitHub.
- [ ] Run `./scripts/prepare_workstation.sh` to validate minimum Go/Node tool
      versions and automatically copy `.env` templates into the backend,
      frontend, and `deploy/` directories.
- [ ] Open the generated environment files (`backend/.env`,
      `frontend/.env.local`, and `deploy/.env`) and replace placeholder values
      with secrets and connection strings that match your local setup.
- [ ] Install any optional tooling called out by the script output (e.g.
      `golangci-lint`, `yt-dlp`, or Docker) so future steps run smoothly.

## 3. Run the stack locally

- [ ] Start PostgreSQL and create a `vidfriends` database (or adjust
      `DATABASE_URL` in `backend/.env` to point at an existing instance).
- [ ] Apply database migrations once they are available. Use
      `go run ./cmd/vidfriends migrate up` from the `backend/` directory, or
      create an issue if the migrations have not yet been added to the
      repository.
- [ ] Launch the Go API with `go run ./cmd/vidfriends serve` and confirm it
      listens on the port configured in your `.env` (default `8080`).
- [ ] Install frontend dependencies with `pnpm install` (or `npm install`) and
      run `pnpm dev` so the Vite dev server starts on port `5173`.
- [ ] Visit `http://localhost:5173`, register a throwaway account, add a friend,
      and share a video link to make sure end-to-end flows succeed.

## 4. Automate container builds

- [ ] Populate `deploy/` with a Docker Compose workflow (or confirm the existing
      definition still matches the stack) so contributors can bring the services
      up with `docker compose up --build`.
- [ ] Provision a GitHub-hosted or self-hosted CI runner with Docker Buildx and
      QEMU so multi-architecture images can be built and pushed.
- [ ] Update the repository's automation (e.g. GitHub Actions) to build and
      publish backend and frontend images on pushes and pull requests.
- [ ] Ensure the build pipeline produces images for both `linux/amd64` and
      `linux/arm64` so developers on Intel and Apple Silicon hardware can run
      the containers.

## 5. Establish team workflows

- [ ] Agree on coding standards, linting rules, and testing expectations for the
      backend and frontend, then document them in the repository.
- [ ] Configure branch protection rules and required status checks for `main`.
- [ ] Groom open issues, prioritize a roadmap, and assign work to collaborators
      so everyone knows what to focus on next.

Completing these steps ensures you have a reliable local environment, understand
how the pieces fit together, and have automation plus team processes ready for
collaborative development.
