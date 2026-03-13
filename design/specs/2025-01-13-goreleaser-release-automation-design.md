# Goreleaser Release Automation Design

**Date:** 2025-01-13
**Author:** Design Session
**Status:** Proposed

## Overview

Design a hybrid release automation system using goreleaser that produces both static and dynamically linked binaries for the service-user microservice. Static binaries provide maximum portability, while dynamic binaries (glibc-specific) offer smaller size and memory efficiency for traditional deployments.

## Requirements

### Functional Requirements
- Support static binary builds for Linux, macOS, and Windows
- Support dynamic binary builds for Linux with multiple glibc versions
- Produce artifacts in industry-standard tar.gz format
- Include configuration example and documentation in each archive
- Generate SHA256 checksums for all artifacts
- Support both amd64 and arm64 architectures

### Platform Support

| Platform | Architectures | Linking | Purpose |
|----------|---------------|---------|---------|
| Linux | amd64, arm64 | Static + Dynamic | Maximum compatibility + efficiency options |
| macOS | amd64, arm64 | Static | Developer workstations |
| Windows | amd64 | Static | Developer workstations |

### Glibc Versions for Dynamic Builds

| Glibc Version | Debian | Ubuntu | Other | Release Era |
|---------------|--------|--------|-------|-------------|
| 2.28 | Debian 10 | Ubuntu 18.04 | CentOS 7+ | 2019 |
| 2.31 | Debian 11 | Ubuntu 20.04 | CentOS 8+ | 2021 |
| 2.36 | Debian 12 | Ubuntu 22.04 | RHEL 9+ | 2023 |
| 2.39 | Debian 13 (testing) | Ubuntu 24.04 | Fedora 40+ | 2024 |

## Design

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Release Process                         │
└─────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   Static Builds        Dynamic Builds         Validation
        │                     │                     │
   ┌────┴────┐         ┌─────┴─────┐              │
   │ Linux   │         │ glibc228  │              │
   │ macOS   │         │ glibc231  │              │
   │ Windows │         │ glibc236  │              │
   └────┬────┘         │ glibc239  │              │
        │             └─────┬─────┘              │
        │                   │                     │
        └───────────────────┴─────────────────────┘
                            │
                    ┌───────┴────────┐
                    │  Artifacts     │
                    │  Archives      │
                    │  Checksums     │
                    │  Signatures    │
                    └────────────────┘
```

### Build Variants

#### Static Binaries

Pure Go builds with `CGO_ENABLED=0`, no external dependencies:

- `service-user_{version}_linux_amd64.tar.gz`
- `service-user_{version}_linux_arm64.tar.gz`
- `service-user_{version}_darwin_amd64.tar.gz`
- `service-user_{version}_darwin_arm64.tar.gz`
- `service-user_{version}_windows_amd64.tar.gz`

**Characteristics:**
- Size: ~50MB per binary
- Dependencies: None (self-contained)
- Compatibility: All Linux distributions (glibc AND musl)
- Use case: Containers, Alpine, unknown environments, maximum portability

#### Dynamic Binaries

Dynamically linked against glibc, built in Debian containers:

- `service-user_{version}_linux_glibc228_amd64.tar.gz`
- `service-user_{version}_linux_glibc228_arm64.tar.gz`
- `service-user_{version}_linux_glibc231_amd64.tar.gz`
- `service-user_{version}_linux_glibc231_arm64.tar.gz`
- `service-user_{version}_linux_glibc236_amd64.tar.gz`
- `service-user_{version}_linux_glibc236_arm64.tar.gz`
- `service-user_{version}_linux_glibc239_amd64.tar.gz`
- `service-user_{version}_linux_glibc239_arm64.tar.gz`

**Characteristics:**
- Size: ~15MB per binary
- Dependencies: Host system's glibc (minimum version per variant)
- Compatibility: Forward-compatible with newer glibc versions
- Use case: Traditional servers, memory efficiency, system integration

### Compatibility Matrix

| Binary | Min Glibc | Compatible Distributions |
|--------|-----------|--------------------------|
| `glibc228` | 2.28 | Debian 10+, Ubuntu 18.04+, CentOS 7+, Fedora 28+ |
| `glibc231` | 2.31 | Debian 11+, Ubuntu 20.04+, CentOS 8+, Fedora 32+ |
| `glibc236` | 2.36 | Debian 12+, Ubuntu 22.04+, RHEL 9+, Fedora 37+ |
| `glibc239` | 2.39 | Debian 13, Ubuntu 24.04+, Fedora 40+ |
| `static` | N/A | **All distributions including Alpine** |

### Release Artifacts

A single release produces 15 artifacts:

```
Release v1.0.0 Assets:

Static Binaries (5):
├── service-user_1.0.0_linux_amd64.tar.gz
├── service-user_1.0.0_linux_arm64.tar.gz
├── service-user_1.0.0_darwin_amd64.tar.gz
├── service-user_1.0.0_darwin_arm64.tar.gz
├── service-user_1.0.0_windows_amd64.tar.gz

Dynamic Binaries (8):
├── service-user_1.0.0_linux_glibc228_amd64.tar.gz
├── service-user_1.0.0_linux_glibc228_arm64.tar.gz
├── service-user_1.0.0_linux_glibc231_amd64.tar.gz
├── service-user_1.0.0_linux_glibc231_arm64.tar.gz
├── service-user_1.0.0_linux_glibc236_amd64.tar.gz
├── service-user_1.0.0_linux_glibc236_arm64.tar.gz
├── service-user_1.0.0_linux_glibc239_amd64.tar.gz
├── service-user_1.0.0_linux_glibc239_arm64.tar.gz

Verification (2):
├── checksums.txt
├── checksums.txt.sig
```

## Implementation

### Goreleaser Configuration

The complete `.goreleaser.yaml` configuration includes:

1. **Static builds** for Linux, macOS, Windows
2. **Dynamic builds** for 4 glibc versions (2.28, 2.31, 2.36, 2.39)
3. **Archive templates** with platform-specific naming
4. **Checksums and signatures** for verification
5. **Build hooks** for preparation steps

See appendix for full configuration.

### Documentation: DYNAMIC_NOTES.md

Each dynamic archive includes `DYNAMIC_NOTES.md` with:
- Explanation of dynamic linking
- Compatible distributions for each glibc version
- How to check system glibc version
- Selection guide for users

## Migration Plan

1. **Phase 1:** Add goreleaser configuration alongside existing workflow
2. **Phase 2:** Test goreleaser locally with `goreleaser release --snapshot`
3. **Phase 3:** Update CI/CD to use goreleaser for releases
4. **Phase 4:** Archive old GitHub Actions workflow

## Trade-offs

| Aspect | Static | Dynamic |
|--------|--------|---------|
| Binary Size | ~50MB | ~15MB |
| Portability | All Linux (glibc + musl) | glibc systems only |
| Memory Usage | Not shared | Shared with other processes |
| Container Use | Ideal | No benefit |
| Traditional Servers | Good | Better (if compatible) |
| Build Complexity | Simple | Requires Docker |
| Dependency Hell | None | Possible version mismatch |

## Future Enhancements

1. Package manager support (deb/rpm/apk)
2. Homebrew tap for macOS
3. Scoop bucket for Windows
4. SBOM generation
5. Auto-discovery release notes

## Appendix: Complete Configuration

```yaml
# .goreleaser.yaml
project_name: service-user

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  # Static builds
  - id: static-linux
    main: ./cmd/main.go
    binary: service-user
    env: [CGO_ENABLED=0]
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags: [-s -w, -extldflags '-static']
    tags: [netgo, osusergo]

  - id: static-darwin
    main: ./cmd/main.go
    binary: service-user
    env: [CGO_ENABLED=0]
    goos: [darwin]
    goarch: [amd64, arm64]
    ldflags: [-s -w]

  - id: static-windows
    main: ./cmd/main.go
    binary: service-user
    env: [CGO_ENABLED=0]
    goos: [windows]
    goarch: [amd64]
    ldflags: [-s -w]

  # Dynamic builds
  - id: dynamic-glibc228
    main: ./cmd/main.go
    binary: service-user
    env: [CGO_ENABLED=1]
    goos: [linux]
    goarch: [amd64, arm64]
    ldflags: [-s -w]
    command: |
      docker run --rm \
        -v "{{ .Env.HOME }}/.cache/go-build:/root/.cache/go-build" \
        -v "$(pwd):/go/src/{{ .ProjectName }}" \
        -w /go/src/{{ .ProjectName }} \
        -e CGO_ENABLED=1 \
        debian:10-slim \
        bash -c "apt-get update -qq && apt-get install -y -qq golang-go && go build -ldflags '-s -w' -o '{{ .Path }}'"

  - id: dynamic-glibc231
    # Similar configuration for Debian 11
    ...

  - id: dynamic-glibc236
    # Similar configuration for Debian 12
    ...

  - id: dynamic-glibc239
    # Similar configuration for Ubuntu 24.04
    ...

archives:
  # Static and dynamic archive configurations
  ...

checksum:
  name_template: 'checksums.txt'

signs:
  - artifacts: all
```

## References

- goreleaser documentation: https://goreleaser.com/
- Static vs dynamic linking: https://golang.org/cmd/link/
- glibc versioning: https://sourceware.org/glibc/wiki/Versioning
