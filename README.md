# service-user

Core identity microservice for the MTAmedia backend ecosystem. Provides user management, authentication, and profile data.

## Features

- **User Management**: Registration, profile management, device tracking
- **Authentication**: JWT-based auth with configurable token expiration
- **OAuth Integration**: Google OAuth support for account linking
- **Security**: Account lockout, password complexity, PIN verification with attempt limiting
- **Event Publishing**: Multi-adapter event system (Redis, Kafka, RabbitMQ)
- **Observer Pattern**: Real-time state change notifications
- **Configurable Hashing**: Argon2 (default), bcrypt, SHA-256 for passwords/PINs

## Quick Start

### Local Development

Build and run the service:

```bash
make build
make run
```

Or run directly:

```bash
go run ./cmd
```

## Architecture

Strict Hexagonal Architecture (ADR-017) with role-based adapter grouping (ADR-012).

```
internal/
├── cmd/                    # Application entrypoints
├── adapter/                # External layer (depends on core)
│   ├── handler/           # Inbound adapters (gRPC handlers)
│   ├── repository/        # Outbound adapters (PostgreSQL)
│   ├── monitoring/        # Outbound adapters (logger, tracer, metrics)
│   └── security/          # security adapters (jwt generator, password hasher, etc)
└── core/                   # Domain layer (no external dependencies)
    ├── domain/            # Pure domain entities
    ├── port/              # Interfaces (inbound & outbound contracts)
    └── service/           # Business logic implementations
```

### Dependency Flow

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
    ├── user_device (M:N) - tracked devices with IP/sessions
    └── user_file (1:N) - uploaded files
```

## Tech Stack

- **Go 1.25.5** with standard library
- **PostgreSQL 18** via pgx
- **gRPC** with Protocol Buffers and gRPC-Gateway
- **Redis** for caching
- **RabbitMQ** for event publishing
- **Kafka** for event streaming
- **Argon2id** for password/PIN hashing (configurable)
- **UUID v7** for public identifiers
- **JWT** for authentication

## Development

### Prerequisites

- Go 1.25.5+
- Docker & Docker Compose
- Make

### Makefile Commands

```bash
make test              # Run all tests
make test verbose      # Run tests with verbose output
make test-cover        # Run tests with coverage and race detection
make test-clean        # Clean test cache and coverage files
make mocks             # Generate mocks using mockery
make build             # Build Linux binary
make release-build     # Build with version injection
make run               # Run the service
make lint              # Run linter (golangci-lint)
make fmt               # Format code
make help              # Show available commands
```

### Database Migrations

See [service-user-db](../service-user-db/README.md) for migration commands.

```bash
cd ../service-user-db
make status            # Check migration status
make update            # Apply migrations
make rollback-count N=1  # Rollback N changesets
```

## Configuration

The service looks for configuration in this order:

1. `--config` flag (highest priority)
2. `/etc/service-user/config.yaml`
3. `./config.yaml` (current directory, fallback)

Example:

```bash
# Install config example
sudo mkdir -p /etc/service-user
sudo cp config.yaml.example /etc/service-user/config.yaml
sudo editor /etc/service-user/config.yaml

# Or use custom location
service-user --config /path/to/custom-config.yaml
```

## Deployment

### Running with systemd

To run the service as a systemd service:

#### 1. Create the service user

```bash
sudo useradd -r -s /bin/false service-user
```

#### 2. Create systemd service file

Create `/etc/systemd/system/service-user.service`:

```ini
[Unit]
Description=Service User gRPC Service
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=service-user
Group=service-user
ExecStart=/usr/local/bin/service-user --config /etc/service-user/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=service-user

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/service-user

# Resource limits
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

#### 3. Install and enable the service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable the service to start on boot
sudo systemctl enable service-user

# Start the service
sudo systemctl start service-user

# Check service status
sudo systemctl status service-user

# View logs
sudo journalctl -u service-user -f
```

## Documentation

Comprehensive documentation is available in the [docs](./docs) directory:

### Architecture & Design

- [Architecture Overview](./docs/architecture/architecture.md)
- [ADR Index](./docs/architecture/adr/README.md)

### Key Architecture Decisions

**Core Architecture:**

- [ADR-001: Go for backend](./docs/architecture/adr/001-programming-language.md)
- [ADR-002: PostgreSQL choice](./docs/architecture/adr/002-database-choice.md)
- [ADR-008: gRPC with Protocol Buffers](./docs/architecture/adr/008-service-communication-protocol.md)
- [ADR-009: Architecture Pattern](./docs/architecture/adr/009-architecture-pattern.md)
- [ADR-017: Strict Hexagonal Architecture](./docs/architecture/adr/017-hexagonal-architecture-strictness.md)
- [ADR-020: Domain File Structure](./docs/architecture/adr/020-core-domain-file-structure.md)

**Database & IDs:**

- [ADR-004: Integer-based internal PKs](./docs/architecture/adr/004-primary-key-strategy.md)
- [ADR-005: UUID v7 for public IDs](./docs/architecture/adr/005-resource-identification.md)

**Security:**

- [ADR-021: Configurable Hashing](./docs/architecture/adr/021-configurable-hashing.md)
- [ADR-033: Account Lockout Threshold](./docs/architecture/adr/033-account-lockout-threshold.md)
- [ADR-037: Password Complexity](./docs/architecture/adr/037-password-complexity-requirements.md)
- [ADR-043: PIN Requirements](./docs/architecture/adr/043-pin-requirements.md)

**Testing:**

- [ADR-024: Testing Strategy](./docs/architecture/adr/024-testing-strategy.md)
- [ADR-025: Table-Driven Tests](./docs/architecture/adr/025-test-structure.md)
- [ADR-047: Mocks Usage Strategy](./docs/architecture/adr/047-mocks-usage-strategy.md)

**Events:**

- [ADR-038: Event Publishing Interface](./docs/architecture/adr/038-event-publishing-interface.md)
- [ADR-039: Event Adapter Redis](./docs/architecture/adr/039-event-adapter-redis.md)
- [ADR-040: Event Adapter Kafka](./docs/architecture/adr/040-event-adapter-kafka.md)
- [ADR-041: Event Adapter RabbitMQ](./docs/architecture/adr/041-event-adapter-rabbitmq.md)

### Technical Specifications

- [TS-01: Database Schema](./docs/technical/technical-specification/TS-01-database-schema.md)

### Project Documentation

- [Service User Documentation Hub](./docs/README.md)
