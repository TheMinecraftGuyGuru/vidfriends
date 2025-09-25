# VidFriends Configuration Templates

This directory collects example environment files for the different ways you can
run VidFriends locally. Copy the template that matches your workflow into the
location each service expects and then customise the values for your
environment.

| Template | Copy to | Purpose |
| --- | --- | --- |
| `backend.env.example` | `backend/.env` | Settings for running the Go API directly on your machine. |
| `frontend.env.example` | `frontend/.env.local` | Vite/React development server configuration. |
| `deploy.env.example` | `deploy/.env` | Shared settings consumed by Docker Compose. |

> **Tip:** The service directories still contain `.env.example` files so
> existing scripts and instructions keep working. Those files are symlinks to the
> templates in this directory, so update the files here when configuration
> options change.
