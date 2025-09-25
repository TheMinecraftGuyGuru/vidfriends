# syntax=docker/dockerfile:1.6
FROM golang:1.22-alpine AS runtime
LABEL org.opencontainers.image.source="https://github.com/${GITHUB_REPOSITORY}"
LABEL org.opencontainers.image.description="VidFriends backend service container image (placeholder)."
WORKDIR /app
COPY backend/README.md /app/README.md
ENV VIDFRIENDS_SERVICE=backend
CMD ["/bin/sh", "-c", "echo VidFriends backend image placeholder"]
