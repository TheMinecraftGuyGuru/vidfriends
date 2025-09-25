# syntax=docker/dockerfile:1.6
FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache build-base

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/vidfriends ./cmd/vidfriends

FROM alpine:3.20 AS runtime
RUN addgroup -S vidfriends && adduser -S -G vidfriends vidfriends
WORKDIR /app

COPY --from=builder /out/vidfriends /usr/local/bin/vidfriends
COPY backend/migrations ./migrations
COPY backend/seeds ./seeds

RUN chown -R vidfriends:vidfriends /app /usr/local/bin/vidfriends
USER vidfriends

ENV VIDFRIENDS_MIGRATIONS=/app/migrations
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/vidfriends"]
CMD ["serve"]
