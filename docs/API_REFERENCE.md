# VidFriends API Reference (Prototype)

This reference documents the HTTP endpoints currently exposed by the VidFriends backend. The service is in an active prototype
state, so behaviors may change quickly and several endpoints still return placeholder data or rely on mocks. All routes are
prefixed with `http://localhost:8080` during local development.

## Authentication

| Method | Path | Status | Notes |
| ------ | ---- | ------ | ----- |
| POST | `/api/v1/auth/signup` | ✅ Implemented | Creates a user and returns session tokens. Requires PostgreSQL migrations and bcrypt-hashed passwords. |
| POST | `/api/v1/auth/login` | ✅ Implemented | Issues session tokens for an existing user. Returns 401 for unknown email or bad password. |
| POST | `/api/v1/auth/refresh` | ✅ Implemented | Exchanges a refresh token for a new session. Fails if the refresh token is missing, expired, or not found. |
| POST | `/api/v1/auth/password-reset` | 🚧 Placeholder | Accepts an email and always responds with `202 Accepted`. No email delivery is wired up yet. |

### Request/response examples

#### Sign up

```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "supersecret"
}
```

Success returns `201 Created` with session tokens:

```json
{
  "tokens": {
    "accessToken": "<JWT>",
    "refreshToken": "<opaque>",
    "expiresAt": "2024-05-01T12:00:00Z"
  }
}
```

#### Refresh session

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "<opaque>"
}
```

Returns `200 OK` with a new `tokens` payload. Use the refresh token from a prior login or signup response.

## Friends

| Method | Path | Status | Notes |
| ------ | ---- | ------ | ----- |
| GET | `/api/v1/friends?user=<id>` | ✅ Implemented | Lists friend requests for the given user. Requires an existing user ID. |
| POST | `/api/v1/friends/invite` | ✅ Implemented | Creates a friend request. Returns `409 Conflict` if one already exists. |
| POST | `/api/v1/friends/respond` | ✅ Implemented | Accepts or blocks a friend request. Supply `action`=`accept` or `block`. |

Example invite payload:

```json
{
  "requesterId": "user-123",
  "receiverId": "user-456"
}
```

Example respond payload:

```json
{
  "requestId": "req-789",
  "action": "accept"
}
```

Responses include the persisted friend request or an error message. All handlers expect valid UUID-style IDs produced by the
backend repositories.

## Videos

| Method | Path | Status | Notes |
| ------ | ---- | ------ | ----- |
| POST | `/api/v1/videos` | ✅ Implemented | Shares a video. Requires `yt-dlp` for metadata lookup; downloads are currently skipped. |
| GET | `/api/v1/videos/feed?user=<id>` | ✅ Implemented | Returns a feed of recent shares for the user and their accepted friends. |

Example share payload:

```json
{
  "ownerId": "user-123",
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
}
```

Successful responses return the stored share with metadata (title, description, thumbnail). Errors are surfaced as JSON with an
`error` field and an appropriate HTTP status.

## Health

| Method | Path | Status | Notes |
| ------ | ---- | ------ | ----- |
| GET | `/healthz` | ✅ Implemented | Returns `200 OK` when the process is healthy. Does not currently check dependencies. |

## Error format

All JSON responses include an `error` field when something goes wrong. For example:

```json
{
  "error": "invalid credentials"
}
```

Use HTTP status codes to determine whether an issue is client-side (4xx) or server-side (5xx). The backend logs provide
additional context—capture those details when filing issues.

## Authentication model

- Access tokens are short-lived JWTs used for API authentication (frontend integration pending).
- Refresh tokens are stored in PostgreSQL via the `sessions` table. Losing the database connection will invalidate future refresh
  attempts.
- Password reset requests are recorded only for observability; actual email delivery and token generation are future work.

## Known gaps

- OAuth/social login, MFA, and device management are not implemented.
- Rate limiting and abuse protections are not configured.
- Object storage uploads are stubbed—video metadata is stored, but no files are persisted.
- Some endpoints may respond with generic error messages while logging detailed diagnostics; improve user-facing error copy as the
  product matures.

Refer back to this document as the prototype evolves, and update it when new endpoints land or behaviors change.
