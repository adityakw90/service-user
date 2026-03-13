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
│   ├── monitoring/        # Outbound adapters (logger, tracer, metrics)
│   └── security/          # security adapters (jwt generator, password hasher, etc)
└── core/                   # Domain layer (no external dependencies)
    ├── domain/            # Pure domain entities
    ├── port/              # Interfaces (inbound & outbound contracts)
    └── service/           # Business logic implementations
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

## Running with systemd

To run the service as a systemd service:

### 1. Create the service user

```bash
sudo useradd -r -s /bin/false service-user
```

### 2. Create systemd service file

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

### 3. Install and enable the service

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

## Documentation

- [Architecture](/docs/architecture/architecture.md)
- [ADR Index](/docs/architecture/adr/README.md)
  - ADR-009: Architecture Pattern
  - ADR-010: Dependency Direction
  - ADR-012: Adapter Grouping
  - ADR-017: Strict Hexagonal Architecture
