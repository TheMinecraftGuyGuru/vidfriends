# Coding Standards & Linting Guide

This guide summarizes the formatting, linting, and stylistic conventions used in the
VidFriends codebase. It complements the Startup and First Steps guides by explaining
how to keep backend Go code and the frontend React + TypeScript project consistent.

## Backend (Go)

### Language level
- **Go version:** The project targets Go 1.21 as declared in `backend/go.mod`.
- **Project layout:** Keep application code inside `internal/`, with HTTP entrypoints
  in `cmd/vidfriends/` and SQL migrations in `migrations/`.

### Formatting
- Run `gofmt` on every Go file. The `scripts/lint.sh` helper script will fail if any
  file under `backend/` is not `gofmt`-clean.
- Use `goimports` (or your editor's import organizer) to group and sort imports; this
  is compatible with the `gofmt` style enforced by the repository.

### Static analysis
- `golangci-lint run` is recommended before opening a pull request. No custom config
  is supplied, so it runs with its default linters. Install instructions are linked
  from `docs/STARTUP.md`.
- Keep handler and domain logic covered by unit tests. Run `go test ./...` or
  `go test -race ./...` for race detection when touching concurrent code.

## Frontend (React + TypeScript)

### Language level
- The frontend uses TypeScript in [strict mode](../frontend/tsconfig.base.json), JSX
  via the React automatic runtime, and targets modern browsers (ES2020).
- Favor functional components with hooks for state and side-effects. Colocate UI
  state and API hooks inside `src/` modules to keep concerns isolated.

### Linting
- ESLint configuration lives in [`frontend/.eslintrc.cjs`](../frontend/.eslintrc.cjs)
  and extends `eslint:recommended`, `plugin:@typescript-eslint/recommended`, and
  `prettier`. This enforces TypeScript best practices while disabling stylistic rules
  that conflict with Prettier.
- Run `pnpm run lint` (or `npm run lint`/`yarn lint` depending on your package
  manager) before committing. The same command is invoked from `scripts/lint.sh`.

### Formatting
- Prettier is expected for formatting `.ts`/`.tsx` files. There is no dedicated
  configuration file, so Prettier's defaults apply. Use `pnpm dlx prettier --write` to
  format sources when editors are not configured to do so automatically.
- ESLint is configured with `eslint-config-prettier` to ensure linting stays compatible
  with Prettier's output.

### Testing
- Keep UI logic covered with Vitest and Testing Library. Run `pnpm test` for unit and
  integration suites, and create component-specific tests alongside source modules.

## Combined lint script

`scripts/lint.sh` orchestrates the repository's lint checks by running `gofmt` against
backend sources and invoking the frontend lint script. Prefer this helper when you need
an end-to-end signal before pushing a branch or opening a pull request.
