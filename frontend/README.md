# VidFriends Frontend

The VidFriends frontend is a React + TypeScript single-page application that lets
users authenticate, manage friend relationships, and share video links.

## Project layout

- `src/` – application components, hooks, and utility modules.
- `public/` – static assets served with the application shell.
- `tests/` – component and integration tests.

## Prerequisites

Install Node.js 18+ and either `pnpm` (preferred) or `npm`.

## Running the development server

1. Copy environment variables:
   ```bash
   cp ../configs/frontend.env.example .env.local
   ```
2. Install dependencies:
   ```bash
   pnpm install
   ```
3. Start the dev server:
   ```bash
   pnpm dev
   ```
   The app is served at `http://localhost:5173` by default.

## Building for production

```bash
pnpm build
```

The output is placed in the `dist/` directory and can be served by any static host.

## Testing

```bash
pnpm test
```

For end-to-end testing, use Playwright or Cypress according to the team's
preference.
