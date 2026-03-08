# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

User Identity Microservice for the MTAmedia backend ecosystem. Early-stage project with Go backend, PostgreSQL database, gRPC communication, and Liquibase migrations.

**Tech Stack:** Go 1.21+ | PostgreSQL 17/18 | Redis | gRPC + gRPC-Gateway | Liquibase | Docker

## Commands

### Database Migrations (service-user-db)

```bash
cd service-user-db

make status      # Check migration status
make update      # Apply migrations
make validate    # Validate changelog files
make update-sql  # Generate SQL without applying
make rollback-count N=1  # Rollback N changesets
make history     # Show migration history
```

Requires environment variables: `DATABASE_URL`, `DATABASE_USER`, `DATABASE_PASSWORD`

### Local Development Infrastructure

```bash
cd service-user

docker-compose up -d  # Starts Redis, PostgreSQL, PgAdmin
# Redis: localhost:6379
# PostgreSQL: localhost:5432 (user: vexa, db: vexa_unittest)
# PgAdmin: localhost:5050 (admin@localhost.id / password)
```

### Go Service (when implemented)

```bash
go test ./...                           # Run all tests
go build -o bin/service-user ./cmd/...  # Build binary
REDIS_HOST=localhost DATABASE_HOST=localhost DATABASE_NAME=vexa_unittest go test ./...  # Run with DB
```

## Architecture

### Directory Structure

- `service-user/` - Go backend service
- `service-user-db/` - Database schema and Liquibase migrations
- `service-user-proto/` - Protocol Buffer definitions
- `service-user-doc/` - Documentation and ADRs

### Data Model

```
user (id, uid, username, email, password, status, timestamps)
    ├── user_profile (1:1) - bio, attributes (JSONB), avatar
    ├── user_pin (1:1) - hashed PIN for security
    ├── user_device (M:N) - tracked devices with IP/sessions
    └── user_file (1:N) - uploaded files
```

### Communication Pattern

```
Client → gRPC-Gateway → gRPC Service → PostgreSQL/Redis
```

### Key ADRs to Review

- ADR-001: Go for backend
- ADR-002: PostgreSQL choice
- ADR-004: Integer-based internal PKs (BigInt)
- ADR-005: UUID v7 for public IDs
- ADR-008: gRPC with Protocol Buffers

## Conventions

### Database
- Table names: Singular, snake_case (e.g., `user`, `user_profile`)
- Primary keys: Internal `id` (BigInt) + Public `uid` (UUID v7)
- Timestamps: All tables use `created_at`, `updated_at`, `deleted_at`

### Code Organization (Go)
- Standard layout: `cmd/`, `internal/`, `pkg/`
- Layered architecture: Handler → Service → Repository
- **Tests: Table-driven using Go's built-in `testing` package**

### Table-Driven Tests (ADR-025)
All unit tests with multiple scenarios must use the table-driven pattern:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "Happy Path", input: "foo", want: "bar", wantErr: false},
        {name: "Error Case", input: "baz", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            // assertions
        })
    }
}
```

### Security
- Passwords: Hashed with argon2id before storage
- PIN codes: Hashed
- PII protection required on `user` and `user_profile` tables

## Key Documentation

- `service-user-doc/README.md` - Documentation hub
- `service-user-doc/architecture/architecture.md` - Architecture overview
- `service-user-doc/technical/technical-specification/TS-01-database-schema.md` - Schema details
- `service-user-db/README.md` - Database setup

## Key ADRs Reference
| ADR | Topic |
|-----|-------|
| ADR-009 | Hexagonal Architecture Pattern |
| ADR-012 | Adapter Directory Structure |
| ADR-017 | Strict Hexagonal Enforcement |
| ADR-020 | Domain File Structure |
| ADR-021 | Configurable Hashing |
| ADR-024 | Testing Strategy |
| ADR-025 | Table-Driven Tests |

## Architecture Rules (from ADRs)

### Strict Hexagonal Architecture (ADR-017)
- Core (`internal/core`) must NOT depend on Adapters
- All inter-layer communication via interfaces in `core/port`
- Adapters convert external data (proto/json) to Domain Entities at boundaries
- Domain entities have no JSON/protobuf tags

### Port Organization
- `core/port/` - Inbound port interfaces (Services)
- `core/port/repository/` - Outbound port interfaces (Repositories)
- `core/port/security/` - Security port interfaces
- `core/port/service/` - Application service interfaces

### Adapter Organization (ADR-012)
- `adapter/handler/` - Inbound adapters (gRPC, HTTP, CLI)
- `adapter/repository/` - Outbound adapters (PostgreSQL implementations)
- `adapter/integration/` - External integrations (logger, mail)
- `adapter/security/` - Security implementations

### Domain File Structure (ADR-020)

`internal/core/domain/` organized by type:

- **`model/`** - Aggregates, Entities, and Value Objects
  - Entities: user.go, device.go, profile.go, file.go, user_pin.go, user_device.go
  - Value Objects: password.go, pin.go, oauth_user_info.go
- **`event/`** - Domain Events (auth_event.go, event_types.go)
- **`errors/`** - Shared domain errors (errors.go)

All domain objects go in `model/`. Import errors from `domain/errors` package.

## Domain Reorganization Pattern

When moving/creating domain files:

1. All domain objects → `model/` (entities + value objects)
2. Domain Events → `event/` (auth_event.go, event_types.go)
3. Domain Errors → `errors/` (errors.go)
4. Import errors from `domain/errors` package (no re-export)
5. Run `go build ./...` after major changes to verify imports
6. Run `go test ./...` to ensure no regressions

### Testing Strategy (ADR-024)
| Layer | Test Type | Mocks | Location |
|-------|-----------|-------|----------|
| Domain | Unit | None | `internal/core/domain/**/*_test.go` |
| Service | Unit | Mock Ports (Repositories) | `internal/core/service/**/*_test.go` |
| Adapters | Unit | Infrastructure mocks (pgxmock) | `internal/adapter/**/*_test.go` |
| Adapters | Integration | None (real infra) | `test/integration/**/*_test.go` |

- **Domain**: Pure unit tests, no mocks. Test real entities directly.
- **Service**: Mock ports (repositories), use real domain objects.
- **Adapters**: Unit tests with infrastructure mocks (pgxmock, miniredis) OR integration tests with real infrastructure.
  - Use `pgxmock` for PostgreSQL mocking in unit tests
  - Use `miniredis` for Redis mocking in unit tests
  - Integration tests in `test/integration/` use real infrastructure (Docker)

### Configurable Hashing (ADR-021)
- Use factory pattern in `adapter/security/factory.go`
- Algorithm selected via env vars: `PASSWORD_HASHER`, `PIN_HASHER`
- Supported: `argon2` (default), `bcrypt`, `sha256`

### gRPC Adapter Pattern

The service uses a centralized gRPC server pattern following the service-access standard:

- `adapter/api/grpc/server.go`: Server lifecycle and service registration
- Handlers receive validator via constructor (DI pattern)
- Middleware handles error conversion via `MakeErrorResponse`
- Response mappers organized by domain entity

Example usage:
```go
grpcServer := grpcadapter.NewServer(
    authService,
    userService,
    deviceService,
    userFileService,
    monitoring,
)
grpcServer.Start(":50051")
```

## Proto Generation

```bash
cd service-user-proto && make go  # Generate Go code before building
```

Required after modifying `.proto` files.

## Dependency Management

- JWT tokens: Use `github.com/golang-jwt/jwt/v5`
- Run `go mod tidy` after adding new dependencies
