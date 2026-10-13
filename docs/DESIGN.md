# Design Notes — Nexus

Why Nexus is built the way it is: the service boundaries, concurrency model, outbox pattern, gRPC contracts, database schema, middleware pipeline, and design decisions.

---

## Architecture

```
┌─────────────┐     HTTP/REST      ┌─────────────┐
│   Client    │ ──────────────────▶ │  Gateway    │
└─────────────┘                    │  :8080      │
                                    └──────┬──────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    ▼                      ▼                      ▼
            ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
            │  Auth Svc     │      │  Users Svc    │      │  Order Svc    │
            │  :5000        │      │  :6000        │      │  :7000        │
            │  HTTP + gRPC  │      │  HTTP         │      │  HTTP         │
            └───────┬───────┘      └───────┬───────┘      └───────┬───────┘
                    │                      │                      │
                    ▼                      ▼                      ▼
            ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
            │  PostgreSQL   │      │  PostgreSQL   │      │  PostgreSQL   │
            │  (auth DB)    │      │  (users DB)   │      │  (orders DB)  │
            └───────┬───────┘      └───────────────┘      └───────────────┘
                    │
                    ▼
            ┌───────────────┐
            │     Redis     │
            │  (sessions,   │
            │  rate limits, │
            │  blacklist)   │
            └───────┬───────┘
                    │
                    ▼
            ┌───────────────┐
            │    Kafka      │
            │  (outbox      │
            │   events)     │
            └───────────────┘
```

### Service boundaries

| Service                                                                               | Responsibility                                                   | Protocols          | Data Store               |
| ------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | ------------------ | ------------------------ |
| **Gateway**                                                                           | Reverse proxy, routing, rate limiting, auth enforcement          | HTTP               | Redis (rate limit)       |
| **Auth**                                                                              | Identity, sessions, passwords, email verification, outbox events | HTTP + gRPC (cmux) | PostgreSQL, Redis, Kafka |
| **Users**                                                                             | Buyer/seller profiles, delegates auth to Auth service            | HTTP               | PostgreSQL               |
| **Shared ([`github.com/ajaka/nexus-shared`](https://github.com/ajaka/nexus-shared))** | Transactional outbox engine, utilities, error defs               | —                  | —                        |

---

## Gateway Service

### Request flow

```
Client Request
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│ Middleware Chain (order matters)                            │
│  1. GenerateRequestID — UUID per request                    │
│  2. AttachScopedLogger — slog with request_id               │
│  3. RateLimit — token bucket per (IP, service)              │
│  4. AuthenticatePrivateRoutes — gRPC call to Auth svc       │
└─────────────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────────────────────────────┐
│ Proxy Handler — path-based routing via service_config.yml   │
│  /api/auth/*   → AUTH_SERVICE  (localhost:5000)             │
│  /api/users/*  → USER_SERVICE  (localhost:6000)             │
│  /api/orders/* → ORDER_SERVICE (localhost:7000)             │
└─────────────────────────────────────────────────────────────┘
```

### Middleware details

**GenerateRequestID** — Sets `X-Request-ID` header (UUID v4) for tracing.

**AttachScopedLogger** — Creates `slog.Logger` with `request_id` attribute, stored in Gin context via `c.Set("logger", logger)`.

**RateLimit** — Redis-backed token bucket (see Rate Limiting below). Key: `limiter:{clientIP:serviceKey}`. Cost from `service_config.yml`.

**AuthenticatePrivateRoutes** — Skips paths under `/api/auth`. Reads `sessionId` cookie, calls Auth gRPC `AuthenticateUser(sessionID)`. On success, sets `X-REQUEST-IDENTITY` header (base64 JSON of `MinimalUserStruct{UserId, Email}`) and `user` in context.

### Rate limiting

Algorithm: **Token bucket** implemented as a Redis Lua script (atomic).

Parameters (in `domain/domain.go`):

- `BUCKET_SIZE` — max tokens
- `REFILL_RATE_PER_SECOND` — tokens added per second

Script logic (in `gateway/internal/cache/cache.go`):

```lua
local tokens = get(tokens_key) or bucket_size
local last = get(last_refill_key) or now
local elapsed = max(0, now - last)
tokens = min(bucket_size, tokens + elapsed * refill_rate)

if tokens >= cost then
    tokens = tokens - cost
    set(tokens_key, tokens), set(last_refill_key, now)
    return {1, floor(tokens), 0}  -- allowed, remaining, retry_after
else
    return {0, floor(tokens), ceil((cost - tokens) / refill_rate)}
end
```

Per-service cost configured in `service_config.yml` (e.g., `AUTH_SERVICE: 10`, `USER_SERVICE: 8`).

### Service discovery

Static YAML configuration (`gateway/configs/service_config.yml`):

```yaml
services:
  AUTH_SERVICE:
    host: "localhost"
    port: 5000
    protocol: "http"
    rate_limiting_cost: 10
```

Loaded at startup via Viper, validated with go-playground/validator.

### Proxy behavior

- Strips `/api` prefix before forwarding
- Preserves query parameters, headers, body, method
- Copies upstream response headers, status, body to client
- Timeouts: HTTP client default (no explicit timeout set — TODO)

---

## Auth Service

### Dual-protocol server (cmux)

Single listener on `:5000`, multiplexed via `soheilhy/cmux`:

```
tcp :5000
    │
    ├─ cmux.HTTP2() ──────▶ gRPC Server (AuthService)
    │
    └─ cmux.HTTP1Fast() ──▶ Gin HTTP Server (/auth/* routes)
```

Runs both servers in goroutines, blocks on `mux.Serve()`.

### REST endpoints (`/auth`)

| Method | Path             | Handler              | Auth Required |
| ------ | ---------------- | -------------------- | ------------- |
| POST   | /register        | HandleRegister       | No            |
| POST   | /login           | HandleLogin          | No            |
| POST   | /logout          | HandleLogout         | Yes           |
| POST   | /refresh         | HandleRefresh        | Yes           |
| POST   | /password/change | HandleChangePassword | Yes           |
| POST   | /password/forgot | HandleForgotPassword | No            |
| POST   | /password/reset  | HandleResetPassword  | No            |
| GET    | /email/verify    | HandleEmailVerify    | No            |
| GET    | /sessions        | HandleListSessions   | Yes           |
| DELETE | /sessions/:id    | HandleRevokeSession  | Yes           |

### gRPC API (`proto/auth/v1/auth.proto`)

```protobuf
service AuthService {
  rpc VerifyUserExists(VerifyUserExistsRequest) returns (VerifyUserExistsResponse);
  rpc AuthenticateUser(AuthenticateUserRequest) returns (AuthenticateUserResponse);
}

message VerifyUserExistsRequest { string id = 1; string email = 2; }
message VerifyUserExistsResponse { bool exists = 1; }

message AuthenticateUserRequest { string sessionID = 1; }
message MinimalUserStruct { string userId = 1; string email = 2; }
message AuthenticateUserResponse { bool allow = 1; MinimalUserStruct user = 2; }
```

**VerifyUserExists** — Called by Users service. Validates user exists, is verified. Returns `exists: true/false`.

**AuthenticateUser** — Called by Gateway. Looks up session in Redis. Returns `allow` + user info if valid.

### Middleware chain (Auth service)

```
Gin Router
    │
    ├─ Recovery
    ├─ AttachScopedLogger (all routes)
    └─ /auth group
         ├─ Validation middleware (per-route: register, login, forgot, reset, session)
         └─ AuthenticationMiddleware (protected routes) — cookie → Redis → user context
```

**Validation middleware** — Uses `go-playground/validator` on request structs, sets validated struct in context (`c.Set("registerRequest", req)`).

**AuthenticationMiddleware** — Reads `sessionId` cookie, calls `cache.GetUser(ctx, sessionId)`, sets `user` and `sessionId` in context.

### Session management

- Stored in Redis: hash at `session:{sessionId}` with fields `user_id`, `email`
- TTL: 24 hours (configurable)
- Device tracking: parsed from `User-Agent` via `ua-parser` (brand, browser, OS)
- Max 5 concurrent sessions per user (oldest evicted on 6th)
- Same device/browser reuses session (rotates token)
- Session ID: UUID v7 without hyphens (`GenerateCleanUUID`)
- Cookie: `HttpOnly`, `Secure` in production, `SameSite=Lax`

### Password management

- **Hashing**: bcrypt default cost
- **History**: last password stored in `password_history` table on change
- **Reset cooldown**: 10 days (`passwordResetCooldown`)
- **Reuse prevention**: compares against current hash + all history entries
- **Length limit**: 50 bytes max (`ValidatePasswordLength`)

### Email verification

1. Registration creates user (`verified=false, active=false`)
2. Generates JWT (HS256, `JWT_EMAIL_SECRET`, `JWT_EMAIL_DURATION` hours)
3. Builds payload: `{email, url: FRONTEND_VERIFICATION_URL?token=...}`
4. Writes to outbox table with event `user.register`
5. Outbox engine publishes to Kafka
6. Consumer (not implemented) sends email
7. User clicks link → `GET /email/verify?token=...`
8. Validates JWT, calls `repo.VerifyUser(email)` → sets `verified=true, active=true`

### Transactional Outbox

**Problem**: Reliably publish events to Kafka without distributed transactions.

**Solution**: Write event to `outbox` table in same DB transaction as business logic. Background worker polls and publishes.

**Schema** (`services/auth/migrations/000005_create_outbox_table.up.sql`):

```sql
CREATE TYPE status_enum AS ENUM ('completed','pending','sending','failed');
CREATE TABLE outbox (
    id SERIAL PRIMARY KEY,
    event TEXT NOT NULL,
    payload BYTEA NOT NULL,
    status status_enum NOT NULL DEFAULT 'pending',
    triggered_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    context TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

**Shared engine** (`shared/outbox/outbox.go`):

- `Engine{repo Repository, producer *kafka.Producer, cfg Config}`
- `Start(ctx)` spawns two goroutines:
  1. `listenToDeliveryReports` — handles Kafka acks, calls `repo.MarkAsProcessed/Failed`
  2. `pollLoop` — every `PollInterval`, `FetchPending(limit)`, `MarkAsSending(ids)`, `Produce` each
- Config validation: `PollInterval > 0`, `BatchSize > 0`, `Topic != ""`

**Repository interface** (`shared/outbox/repository.go`):

```go
type Repository interface {
    FetchPending(ctx, limit) ([]Record, error)
    MarkAsSending(ctx, ids) error
    MarkAsProcessed(ctx, id) error
    MarkAsFailed(ctx, id, reason) error
}
```

Auth service implements this in `repositories/repository.go` using `pgx`.

**Events emitted**:

- `user.register` — payload: `{email, url}` (email verification)
- `user.password.reset` — payload: `{email, url}` (password reset)

---

## Users Service

Minimal implementation demonstrating service-to-service gRPC call.

### gRPC client call

```go
conn, _ := grpc.NewClient("localhost:5000", insecure.NewCredentials())
authClient := authv1.NewAuthServiceClient(conn)

res, _ := authClient.VerifyUserExists(ctx, &authv1.VerifyUserExistsRequest{
    Email: email,
    Id:    id.String(),
})
```

### Data model

```go
type User struct {
    Id          uuid.UUID
    FirstName   string
    LastName    string
    MiddleName  string
    PhoneNumber string
    Type        string  // 'buyer', 'seller', 'both'
    Country     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Buyer/Seller profiles in separate tables (migrations 000002, 000003). See `services/users/migrations/` for full schema.

---

## Shared Library (nexus-shared)

Module: [`github.com/ajaka/nexus-shared`](https://github.com/ajaka/nexus-shared)

### Packages

| Package  | Purpose                                                                   |
| -------- | ------------------------------------------------------------------------- |
| `outbox` | Transactional outbox engine + Repository interface                        |
| `utils`  | `EncodeBase64`, `DecodeBase64`, `Stringify` (JSON), `Parse` (base64 JSON) |
| `err`    | `ErrInvalidPollDuration`, `ErrInvalidBatchSize`, `ErrKafkaTopicEmpty`     |

### Outbox engine usage (Auth service)

```go
// In auth service initialization:
repo := repositories.InitRepository(pool)  // implements outbox.Repository
engine, _ := outbox.NewEngine(repo, kafkaBrokers, outbox.Config{
    PollInterval: 5 * time.Second,
    BatchSize:    100,
    Topic:        "auth-events",
}, logger)
engine.Start(ctx)
```

---

## Configuration

### Gateway

Environment variables (`.env` loaded via godotenv):

- `SERVER_ADDR` (e.g., `:8080`)
- `ENVIRONMENT` (`development`|`production`)
- `REDIS_ADDR`
- `REDIS_PASSWORD` (optional)

Service config: `gateway/configs/service_config.yml` (see above).

### Auth Service

Environment variables:

- `SERVER_ADDR` (`:5000`)
- `DATABASE_URL` (PostgreSQL DSN)
- `REDIS_ADDR`, `REDIS_PASSWORD`
- `KAFKA_BROKERS`
- `JWT_EMAIL_SECRET`, `JWT_EMAIL_DURATION` (hours)
- `FRONTEND_VERIFICATION_URL`
- `PRODUCTION` (bool)

### Users Service

Environment variables:

- `SERVER_ADDR` (`:6000`)
- `DATABASE_URL`
- `AUTH_SERVICE_ADDR` (`localhost:5000`)

---

## Design Decisions & Trade-offs

| Decision                                    | Rationale                                           | Trade-off                                           |
| ------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- |
| **cmux for dual-protocol**                  | Single port, simpler deployment                     | Slightly more complex server setup                  |
| **Static YAML service discovery**           | Simplicity for learning project                     | No dynamic scaling, manual updates                  |
| **gRPC for internal, HTTP for gateway**     | Performance + contracts internally; REST externally | Two protocols to maintain                           |
| **Transactional outbox**                    | Reliability without 2PC                             | Added latency (poll interval), eventual consistency |
| **Redis token bucket (Lua)**                | Atomic, distributed, no local state                 | Redis dependency, Lua script maintenance            |
| **JWT for email tokens, sessions in Redis** | Short-lived tokens stateless; sessions revocable    | Two auth mechanisms                                 |
| **bcrypt for passwords**                    | Standard, tunable cost                              | CPU-intensive (intentional)                         |
| **uuidv7 for IDs**                          | Time-ordered, better index locality                 | Requires PostgreSQL 16+ or extension                |
| **Shared outbox library**                   | Reusable across services                            | Couples services to shared module version           |

---

## Learning Outcomes Demonstrated

- ✅ Microservice decomposition & boundaries
- ✅ API Gateway pattern (routing, middleware, proxy)
- ✅ gRPC service definitions & code generation
- ✅ Dual-protocol servers (HTTP + gRPC via cmux)
- ✅ Authentication & authorization patterns (Redis-backed sessions, email/reset JWT tokens, gRPC verification)
- ✅ Distributed rate limiting (Redis token bucket)
- ✅ Transactional outbox for event-driven architecture
- ✅ Database migration management (versioned SQL)
- ✅ Shared library design (outbox, utils, errors)
- ✅ Structured logging (slog)
- ✅ Configuration management (env + YAML)
- ✅ Middleware pipeline design
- ✅ Service-to-service communication
- ✅ Error handling & propagation

---
