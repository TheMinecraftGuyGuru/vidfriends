# First steps for the VidFriends project

These steps help a new contributor or maintainer get oriented and prepare the
project for collaborative development.

## 1. Understand the architecture

- [x] Read the [repository overview](../README.md) to understand the backend,
      frontend, and deployment directories.
- [x] Skim `backend/README.md` and `frontend/README.md` (create them if missing)
      to confirm how each service is launched and tested.
- [ ] Review `deploy/docker-compose.yml` and the `migrations/` directory so you
      understand how the services interact and how the database evolves.

## 2. Prepare your development workstation

- [x] Clone the repository and create a personal fork if you plan to open pull
      requests.
- [x] Follow the [Startup Guide](STARTUP.md) to install Go, Node.js, pnpm or
      npm, PostgreSQL, and Docker (if you will use the container workflow). Run
      `./scripts/prepare_workstation.sh` to verify minimum tool versions and
      copy example environment files automatically.
- [x] Review the generated `.env` files (`backend/.env`, `frontend/.env.local`,
      and `deploy/.env`) and edit them with secrets that match your local
      environment.

## 3. Run the stack locally

- [ ] Start PostgreSQL and apply database migrations via
      `go run ./cmd/vidfriends migrate up`.
- [ ] Launch the Go API with `go run ./cmd/vidfriends serve` and verify it
      listens on port 8080.
- [ ] Install frontend dependencies with `pnpm install` (or `npm install`) and
      run the development server using `pnpm dev`.
- [ ] Visit `http://localhost:5173`, create a test account, and confirm you can
      add friends and share a video link end-to-end.

## 4. Automate container builds

- [ ] Provision a GitHub self-hosted runner with Docker Buildx and QEMU
      emulation enabled so that multi-architecture images can be produced.
- [ ] Configure the repository's CI workflows to trigger container builds on
      pushes and pull requests, using the self-hosted runner to build and push
      images.
- [ ] Ensure that the automation publishes containers for both x86_64 and arm64
      so developers can run the stack on laptops, desktops, and ARM-based
      devices.

## 5. Establish team workflows

- [ ] Decide on coding standards, linting rules, and testing requirements for
      both the backend and frontend. Document them in the repository.
- [ ] Set up branch protection rules and required status checks for the main
      branch.
- [ ] Review open issues, prioritize a roadmap, and assign work to
      collaborators.

With these steps complete you will have a reliable local environment, automated
multi-architecture container builds, and clear team practices that make
contributing to VidFriends efficient and predictable.
