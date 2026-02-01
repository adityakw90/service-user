# service-user

Core identity microservice for the MTAmedia backend ecosystem. Provides user management, authentication, and profile data.

## Architecture

Strict Hexagonal Architecture (ADR-017) with role-based adapter grouping (ADR-012).

```
internal/
├── cmd/                    # Application entrypoints
├── adapter/                # External layer (depends on core)
│   ├── handler/           # Inbound adapters (gRPC handlers)
│   ├── repository/        # Outbound adapters (PostgreSQL)
│   └── integration/       # External integrations (logger)
└── core/                   # Domain layer (no external dependencies)
    ├── domain/            # Pure domain entities
    ├── port/              # Interfaces (inbound & outbound contracts)
    ├── service/           # Business logic implementations
    └── security/          # Cryptographic operations
```

## Dependency Flow

```
cmd/server/main.go (composition root)
    ↓ injects
adapter/handler/grpc → core/port (interface) → core/service → core/port (interface) → adapter/repository
```

**Rule**: Source code dependencies only point inward. Core knows nothing about adapters.

## Key Principles (ADR-017)

- **Strict Layer Separation**: Core does not import adapters
- **Mandatory Interfaces**: All inter-layer communication via `core/port`
- **DTO at Boundaries**: Adapters convert external data to domain entities
- **No Domain Coupling**: Domain entities have no JSON tags or framework dependencies

## Data Model

```
user (id, uid, username, email, password, status, timestamps)
    ├── user_profile (1:1) - bio, attributes (JSONB), avatar
    ├── user_pin (1:1) - hashed PIN for security
    └── user_device (M:N) - tracked devices with IP/sessions
```

## Tech Stack

- **Go 1.21+** with standard library
- **PostgreSQL 17/18** via pgx
- **gRPC** with Protocol Buffers
- **Redis** for caching
- **Argon2id** for password/PIN hashing
- **UUID v7** for public identifiers

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 17/18

### Build & Test

```bash
go build ./...
go test ./...
```

## Documentation

- [Architecture](/docs/architecture/architecture.md)
- [ADR Index](/docs/architecture/adr/README.md)
  - ADR-009: Architecture Pattern
  - ADR-010: Dependency Direction
  - ADR-012: Adapter Grouping
  - ADR-017: Strict Hexagonal Architecture
