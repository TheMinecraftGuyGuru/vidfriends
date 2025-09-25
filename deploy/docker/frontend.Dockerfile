# syntax=docker/dockerfile:1.6
FROM node:20-alpine AS builder

WORKDIR /app
ENV PNPM_HOME=/pnpm
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY frontend/pnpm-lock.yaml frontend/package.json ./
RUN pnpm install --frozen-lockfile

COPY frontend/ ./ \
    --exclude=frontend/dist \
    --exclude=frontend/node_modules \
    --exclude=frontend/.env* \
    --exclude=frontend/.vite

RUN pnpm build

FROM nginx:1.27-alpine AS runtime
COPY --from=builder /app/dist /usr/share/nginx/html
COPY deploy/docker/frontend.nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
