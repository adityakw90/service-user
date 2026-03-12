# Service User {VERSION}

## Quick Install

### For Debian, Ubuntu, Fedora, RHEL, CentOS:
```bash
VERSION={VERSION}
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
else
  ARCH="amd64"
fi

wget https://github.com/adityakw90/service-user/releases/download/v${VERSION}/service-user-${VERSION}-linux-${ARCH}
wget https://github.com/adityakw90/service-user/releases/download/v${VERSION}/checksums.txt
sha256sum --ignore-missing -c checksums.txt

chmod +x service-user-${VERSION}-linux-${ARCH}
sudo mv service-user-${VERSION}-linux-${ARCH} /usr/local/bin/service-user
```

### For Alpine Linux:
```bash
VERSION={VERSION}
ARCH=$(uname -m)
if [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
else
  ARCH="amd64"
fi

wget https://github.com/adityakw90/service-user/releases/download/v${VERSION}/service-user-${VERSION}-linux-musl-${ARCH}
wget https://github.com/adityakw90/service-user/releases/download/v${VERSION}/checksums.txt
sha256sum --ignore-missing -c checksums.txt

chmod +x service-user-${VERSION}-linux-musl-${ARCH}
sudo mv service-user-${VERSION}-linux-musl-${ARCH} /usr/local/bin/service-user
```

## Configuration

```bash
sudo mkdir -p /etc/service-user
sudo cp config.yaml.example /etc/service-user/config.yaml
sudo editor /etc/service-user/config.yaml
```

## Running with systemd

```bash
# Create service user
sudo useradd -r -s /bin/false service-user

# Install service file
sudo cp service-user.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now service-user
```

## Version Information

```bash
service-user --version
```

## Binary Selection Guide

| Distro | Use Binary |
|--------|------------|
| Debian, Ubuntu, Fedora, RHEL, CentOS | `service-user-{VERSION}-linux-{ARCH}` |
| Alpine Linux | `service-user-{VERSION}-linux-musl-{ARCH}` |
| Docker (scratch/minimal) | `service-user-{VERSION}-linux-musl-{ARCH}` |

## Database Migrations

Migrations are managed separately via the `service-user-db` repository. Before starting the service:

1. Obtain migrations from `service-user-db`
2. Run migrations using Liquibase or the provided migration scripts

For more information, see: https://github.com/adityakw90/service-user-db