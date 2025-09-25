# syntax=docker/dockerfile:1.6
FROM python:3.12-alpine

# Install the yt-dlp CLI so the backend can download video metadata locally.
RUN pip install --no-cache-dir yt-dlp==2023.11.14

# Copy development seed data and SQL migrations used for onboarding.
WORKDIR /app
COPY backend/seeds/dev_seed.sql seeds/dev_seed.sql
COPY backend/migrations migrations

# Default command copies the binary and seed file into bind-mounted volumes.
# The Compose override redefines target directories via environment variables.
ENV YTDLP_EXPORT_DIR=/export/bin \
    SEED_EXPORT_DIR=/export/seeds

ENTRYPOINT ["/bin/sh", "-c"]
CMD ["\\
set -eu\\n\
mkdir -p \"$${YTDLP_EXPORT_DIR}\" \"$${SEED_EXPORT_DIR}\"\\n\
cp /usr/local/bin/yt-dlp \"$${YTDLP_EXPORT_DIR}/yt-dlp\"\\n\
chmod +x \"$${YTDLP_EXPORT_DIR}/yt-dlp\"\\n\
cp /app/seeds/dev_seed.sql \"$${SEED_EXPORT_DIR}/dev_seed.sql\"\\n\
rm -rf \"$${SEED_EXPORT_DIR}/migrations\"\\n\
cp -r /app/migrations \"$${SEED_EXPORT_DIR}/migrations\"\\n\
" ]
