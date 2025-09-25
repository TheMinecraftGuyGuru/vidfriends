# syntax=docker/dockerfile:1.6
FROM node:20-alpine
LABEL org.opencontainers.image.source="https://github.com/${GITHUB_REPOSITORY}"
LABEL org.opencontainers.image.description="VidFriends frontend application container image (placeholder)."
WORKDIR /app
COPY frontend/README.md /app/README.md
ENV VIDFRIENDS_SERVICE=frontend
CMD ["/bin/sh", "-c", "echo VidFriends frontend image placeholder"]
