# Nexus

A microservices learning project demonstrating gateway routing, dual-protocol services, transactional outbox, and service-to-service gRPC communication — all in Go.

**Work in progress.** This is a learning exercise, not production-ready software.

---

## At a glance

| Aspect        | Detail                                                                                           |
| ------------- | ------------------------------------------------------------------------------------------------ |
| Language      | Go 1.27+                                                                                         |
| Services      | `gateway`, `auth`, `users`                                                                       |
| Shared lib    | [`github.com/ajaka/nexus-shared`](https://github.com/ajaka/nexus-shared) (outbox, utils, errors) |
| Protocols     | HTTP/REST, gRPC (cmux for dual-protocol)                                                         |
| Databases     | PostgreSQL (per service)                                                                         |
| Cache / Queue | Redis (sessions, rate limiting, blacklist), Kafka (outbox events)                                |
| Auth          | Session cookies, gRPC verification                                                               |
| Rate limiting | Redis-backed token bucket                                                                        |

---

## Architecture

```
Client → Gateway (:8080) → Auth Service (:5000, HTTP+gRPC)
                          ↘ Users Service (:6000, HTTP)
```

- **Gateway** — Reverse proxy, request routing, rate limiting, session validation via auth gRPC
- **Auth** — Registration, login, sessions, password management, email verification, outbox events to Kafka
- **Users** — Buyer/seller profiles, calls auth gRPC `VerifyUserExists`
- **Shared** — Transactional outbox pattern (reliable Kafka publishing), utilities, error definitions

---

## Service endpoints

### Gateway (public entry point)

All routes proxied to upstream services under `/api`:

- `POST /api/auth/*` → Auth service
- `GET  /api/users/*` → Users service

### Auth service (`/auth`)

- `POST /register` — User registration (triggers email verification via outbox)
- `POST /login` — Login, returns session cookie
- `POST /logout` — Revoke session
- `POST /refresh` — Token refresh
- `POST /password/change` — Change password (with history check)
- `POST /password/forgot` — Request reset email
- `POST /password/reset` — Reset password
- `GET  /email/verify` — Verify email via token
- `GET  /sessions` — List active sessions
- `DELETE /sessions/:id` — Revoke specific session

### gRPC (Auth service, port 5000)

- `VerifyUserExists(id, email)` — Used by Users service
- `AuthenticateUser(sessionID)` — Used by Gateway for session validation

---

## Project layout

```
nexus/
├── proto/auth/v1/auth.proto        # Shared gRPC contracts
├── gateway/                        # API Gateway
│   ├── cmd/api/main.go
│   ├── internal/{server,routes,middlewares,cache,configs,grpc,common,domain}
│   └── configs/service_config.yml  # Upstream service definitions
├── services/
│   ├── auth/                       # Core auth service
│   │   ├── cmd/api/main.go
│   │   ├── internal/{server,handlers,grpc_server,models,repositories,store,cache,middlewares,configs,errs}
│   │   └── migrations/             # 6 SQL migrations
│   └── users/                      # User profiles (buyer/seller)
│       ├── cmd/api/main.go
│       ├── internal/{server,models}
│       └── migrations/             # 4 SQL migrations
├── shared/                         # Separate module: [`github.com/ajaka/nexus-shared`](https://github.com/ajaka/nexus-shared)
│   ├── outbox/                     # Transactional outbox engine
│   ├── utils/                      # Base64/JSON helpers
│   └── err/                        # Shared error definitions
├── scripts/migrations.sh           # Migration runner
└── dev-artifacts/                  # Design notes (gitignored)
```

---

## Documentation

- [Design notes](docs/DESIGN.md) — Architecture, concurrency, outbox pattern, gRPC contracts, database schema, middleware pipeline, configuration, and design decisions

---

## Status

**Work in progress.** Working today: gateway routing + middleware chain, auth service (HTTP + gRPC via cmux), user registration/login/sessions, password management with history, email verification flow, transactional outbox to Kafka, service-to-service gRPC calls, Redis-backed rate limiting and sessions, database migrations.

---

## License

MIT
