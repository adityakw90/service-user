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
| 2.27 | - | Ubuntu 18.04 | - | 2018 |
| 2.28 | Debian 10 | - | CentOS 7+ | 2019 |
| 2.31 | Debian 11 | Ubuntu 20.04 | CentOS 8+ | 2021 |
| 2.35 | - | Ubuntu 22.04 | - | 2022 |
| 2.36 | Debian 12 | - | RHEL 9+ | 2023 |
| 2.38 | Debian 13 (testing) | - | - | 2024 |
| 2.39 | - | Ubuntu 24.04 | Fedora 40+ | 2024 |

**Note:** We build on glibc 2.28, 2.31, 2.36, and 2.39 to provide coverage from Debian 10 through Ubuntu 24.04.

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

Dynamically linked against glibc, built in specific distribution containers:

- `service-user_{version}_linux_glibc228_amd64.tar.gz` (Debian 10)
- `service-user_{version}_linux_glibc228_arm64.tar.gz`
- `service-user_{version}_linux_glibc231_amd64.tar.gz` (Debian 11)
- `service-user_{version}_linux_glibc231_arm64.tar.gz`
- `service-user_{version}_linux_glibc236_amd64.tar.gz` (Debian 12)
- `service-user_{version}_linux_glibc236_arm64.tar.gz`
- `service-user_{version}_linux_glibc239_amd64.tar.gz` (Ubuntu 24.04)
- `service-user_{version}_linux_glibc239_arm64.tar.gz`

**Characteristics:**
- Size: ~15MB per binary (estimated)
- Dependencies: Host system's glibc (minimum version per variant)
- Compatibility: Forward-compatible with newer glibc versions
- Use case: Traditional servers, memory efficiency, system integration

**Build platforms:**
- glibc228: Built on Debian 10 (buster-slim)
- glibc231: Built on Debian 11 (bullseye-slim)
- glibc236: Built on Debian 12 (bookworm-slim)
- glibc239: Built on Ubuntu 24.04

### Compatibility Matrix

| Binary | Min Glibc | Runs On Systems With glibc ≥ |
|--------|-----------|-----------------------------|
| `glibc228` | 2.28 | Debian 10+, Ubuntu 20.04+, CentOS 8+, Fedora 32+ |
| `glibc231` | 2.31 | Debian 11+, Ubuntu 20.04+, CentOS 8+, Rocky 9+ |
| `glibc236` | 2.36 | Debian 12+, Ubuntu 22.04+, RHEL 9+, Fedora 37+ |
| `glibc239` | 2.39 | Debian 13+, Ubuntu 24.04+, Fedora 40+ |
| `static` | N/A | **All distributions including Alpine** |

**Key points:**
- Each binary requires AT LEAST the minimum glibc version shown
- Forward compatible: glibc228 binary runs on systems with glibc 2.28, 2.31, 2.36, 2.39+
- Backward incompatible: glibc236 binary will NOT run on Debian 10 (glibc 2.28)
- For Ubuntu 18.04 (glibc 2.27), use the static binary

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

Verification Files (2):
├── checksums.txt (SHA256 hashes of all artifacts)
├── checksums.txt.sig (GPG signature of checksums.txt)
```

**Note:** Ubuntu 18.04 LTS (glibc 2.27) users should use the static binary as glibc 2.28 is the minimum for dynamic builds.

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

See Appendix B for the complete template.

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

## Appendix

### Appendix A: DYNAMIC_NOTES.md Template

```markdown
# Dynamic Binary Notes

## What is "dynamic"?

The `*_linux_glibc*_*.tar.gz` binaries are dynamically linked against glibc.
This means they require a compatible glibc version on the host system.

## Benefits of Dynamic Binaries

- ✅ Smaller binary size (~15MB vs ~50MB for static)
- ✅ Shared memory usage when running multiple services on same host
- ✅ System-level security updates via glibc patches
- ✅ Better integration with system monitoring tools

## When to Use Dynamic vs Static

### Use Dynamic When:
- Deploying to traditional Linux servers (not containers)
- Running multiple services on the same host
- Memory efficiency is important
- You control the deployment environment

### Use Static When:
- Running in containers (Docker, Kubernetes)
- Deploying to Alpine Linux (uses musl, not glibc)
- Uncertain about target system
- Maximum portability needed

## Available Variants

### glibc228 (Debian 10)
- **Minimum required:** glibc 2.28 or higher
- **Runs on systems with:** glibc ≥2.28 (forward compatible)
- **Compatible distributions:**
  - Debian 10, 11, 12, 13+
  - Ubuntu 20.04, 22.04, 24.04+
  - CentOS 8, Stream 9
  - Fedora 32+
- **Build platform:** Debian 10 (buster)
- **Use when:** You need to support Debian 10 or want widest compatibility

### glibc231 (Debian 11)
- **Minimum required:** glibc 2.31 or higher
- **Runs on systems with:** glibc ≥2.31 (forward compatible)
- **Compatible distributions:**
  - Debian 11, 12, 13+
  - Ubuntu 20.04, 22.04, 24.04+
  - CentOS 8, Stream 9
  - Fedora 32+
- **Build platform:** Debian 11 (bullseye)
- **Use when:** You're on 2020+ distributions (but not Debian 10)

### glibc236 (Debian 12)
- **Minimum required:** glibc 2.36 or higher
- **Runs on systems with:** glibc ≥2.36 (forward compatible)
- **Compatible distributions:**
  - Debian 12, 13+
  - Ubuntu 22.04, 24.04+
  - RHEL 9, Rocky 9, Alma 9
  - Fedora 37+
- **Build platform:** Debian 12 (bookworm)
- **Use when:** You're on 2023+ distributions

### glibc239 (Ubuntu 24.04)
- **Minimum required:** glibc 2.39 or higher
- **Compatible distributions:**
  - Ubuntu 24.04 LTS (Noble Numbat) and later
  - Fedora 40 and later
  - Debian 13+ (forward compatible)
- **Build platform:** Ubuntu 24.04
- **Use when:** You're on the latest distributions

## How to Check Your glibc Version

```bash
# Method 1: Using ldd
ldd --version | head -1
# Output examples:
# ldd (Debian GLIBC 2.36-5)       ← Your glibc is 2.36
# ldd (Ubuntu GLIBC 2.39-0ubuntu8) ← Your glibc is 2.39

# Method 2: Using libc.so directly
/lib/x86_64-linux-gnu/libc.so.6
# Output: GNU C Library (Debian GLIBC 2.36-5) stable release version 2.36
```

## Quick Selection Guide

| Your glibc Version | Use Binary |
|--------------------|------------|
| 2.28 - 2.30 | `glibc228` |
| 2.31 - 2.35 | `glibc231` |
| 2.36 - 2.38 | `glibc236` |
| 2.39+ | `glibc239` |
| Alpine (musl) | Use static (`*_linux_*.tar.gz`) |
| Not sure | Use static (`*_linux_*.tar.gz`) |

## Distribution to Binary Mapping

| Distribution | Version | Use Binary |
|--------------|---------|------------|
| Debian | 10 (Buster) | `glibc228` |
| Debian | 11 (Bullseye) | `glibc231` |
| Debian | 12 (Bookworm) | `glibc236` |
| Debian | 13 (Trixie) | `glibc239` |
| Ubuntu | 18.04 LTS | **Use static** (glibc 2.27) |
| Ubuntu | 20.04 LTS | `glibc231` |
| Ubuntu | 22.04 LTS | `glibc236` |
| Ubuntu | 24.04 LTS | `glibc239` |
| CentOS | 7 | `glibc228` |
| CentOS | 8 | `glibc231` |
| RHEL / Rocky / Alma | 9 | `glibc236` |
| Alpine | Any | `*_linux_*.tar.gz` (static) |

## Installation Example

```bash
# Check your glibc version first
ldd --version | head -1

# Example output: ldd (Debian GLIBC 2.36-5)
# You have glibc 2.36, use glibc236 binary

# Download and install
wget https://github.com/yourorg/service-user/releases/download/v1.0.0/service-user_1.0.0_linux_glibc236_amd64.tar.gz
tar -xzf service-user_1.0.0_linux_glibc236_amd64.tar.gz
./service-user --version

# Verify dynamic linking
ldd ./service-user
# Output will show: libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6
```

## Troubleshooting

### "version 'GLIBC_2.xx' not found"

This error means the binary requires a newer glibc version than your system has.

**Solution:** Either:
1. Use a binary with a lower glibc requirement
2. Use the static binary (works everywhere)
3. Upgrade your system

### "No such file or directory" on execution

This can occur on ARM systems if you downloaded the amd64 binary (or vice versa).

**Solution:** Verify your architecture:
```bash
uname -m
# x86_64 → use amd64 binaries
# aarch64 → use arm64 binaries
```
```

### Appendix B: Goreleaser Configuration

The complete `.goreleaser.yaml` configuration for this design:

```yaml
# .goreleaser.yaml
project_name: service-user

before:
  hooks:
    - go mod tidy
    - go generate ./...
    - cp DYNAMIC_NOTES.md . || echo "DYNAMIC_NOTES.md not found, skipping"

builds:
  # ═══════════════════════════════════════════════════════════
  # STATIC BUILDS (Portability focus)
  # ═══════════════════════════════════════════════════════════

  - id: static-linux
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=0
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -extldflags '-static'
    tags:
      - netgo
      - osusergo
    flags:
      - -trimpath

  - id: static-darwin
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath

  - id: static-windows
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=0
    goos:
      - windows
    goarch:
      - amd64
    ldflags:
      - -s -w
    flags:
      - -trimpath

  # ═══════════════════════════════════════════════════════════
  # DYNAMIC BUILDS (glibc-specific)
  # ═══════════════════════════════════════════════════════════
  # Note: Dynamic builds require CGO_ENABLED=1 and appropriate build environment
  # These builds should be run in CI/CD with the target distribution containers
  # The configurations below assume the build host has the required toolchains

  # glibc 2.28 (Debian 10 Buster) - Build on Debian 10 or equivalent
  - id: dynamic-glibc228
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=1
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath

  # glibc 2.31 (Debian 11 Bullseye) - Build on Debian 11 or equivalent
  - id: dynamic-glibc231
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=1
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath

  # glibc 2.36 (Debian 12 Bookworm) - Build on Debian 12 or equivalent
  - id: dynamic-glibc236
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=1
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath

  # glibc 2.39 (Ubuntu 24.04) - Build on Ubuntu 24.04 or equivalent
  - id: dynamic-glibc239
    main: ./cmd/main.go
    binary: service-user
    env:
      - CGO_ENABLED=1
    goos:
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
    flags:
      - -trimpath

archives:
  # Static archives - simple naming
  - id: static
    builds:
      - static-linux
      - static-darwin
      - static-windows
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
    format: tar.gz
    files:
      - src: config.yaml.example
        dst: .
      - src: LICENSE
        dst: .
      - src: README.md
        dst: .

  # Dynamic archives - glibc version in filename
  - id: dynamic-glibc228
    builds:
      - dynamic-glibc228
    name_template: "{{ .ProjectName }}_{{ .Version }}_linux_glibc228_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
    format: tar.gz
    files:
      - src: config.yaml.example
        dst: .
      - src: LICENSE
        dst: .
      - src: README.md
        dst: .
      - src: DYNAMIC_NOTES.md
        dst: .

  - id: dynamic-glibc231
    builds:
      - dynamic-glibc231
    name_template: "{{ .ProjectName }}_{{ .Version }}_linux_glibc231_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
    format: tar.gz
    files:
      - src: config.yaml.example
        dst: .
      - src: LICENSE
        dst: .
      - src: README.md
        dst: .
      - src: DYNAMIC_NOTES.md
        dst: .

  - id: dynamic-glibc236
    builds:
      - dynamic-glibc236
    name_template: "{{ .ProjectName }}_{{ .Version }}_linux_glibc236_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
    format: tar.gz
    files:
      - src: config.yaml.example
        dst: .
      - src: LICENSE
        dst: .
      - src: README.md
        dst: .
      - src: DYNAMIC_NOTES.md
        dst: .

  - id: dynamic-glibc239
    builds:
      - dynamic-glibc239
    name_template: "{{ .ProjectName }}_{{ .Version }}_linux_glibc239_{{ .Arch }}{{ if .Arm }}v{{ .Arm }}{{ end }}"
    format: tar.gz
    files:
      - src: config.yaml.example
        dst: .
      - src: LICENSE
        dst: .
      - src: README.md
        dst: .
      - src: DYNAMIC_NOTES.md
        dst: .

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

signs:
  - artifacts: all
  signature: "${artifact}.sig"
  cmd: gpg
  args:
    - "--output"
    - "${signature}"
    - "--detach-sig"
    - "${artifact}"

release:
  draft: false
  prerelease: auto
  mode: replace
  header: |
    ## Release {{ .Tag }} ({{ .Date }})
  footer: |
    ## What's Changed
    Full Changelog: https://github.com/{{ .Env.GITHUB_REPOSITORY }}/compare/{{ .PreviousTag }}...{{ .Tag }}
  extra_files:
    - glob: ./config.yaml.example

changelog:
  disable: false
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'
      - '^build:'
      - '^chore:'
      - '^refactor:'

dist: dist
```

**Important Implementation Notes:**

1. **Dynamic builds require proper build environment:**
   - The configuration above assumes builds run in appropriate distribution containers
   - For CI/CD, use matrix builds with different Docker images:
     ```yaml
     # Example GitHub Actions matrix
     strategy:
       matrix:
         include:
           - build: static-linux
             image: ubuntu:22.04
           - build: dynamic-glibc228
             image: debian:10-slim
           - build: dynamic-glibc231
             image: debian:11-slim
           - build: dynamic-glibc236
             image: debian:12-slim
           - build: dynamic-glibc239
             image: ubuntu:24.04
     ```

2. **Cross-architecture support:**
   - Use QEMU/binfmt_misc for cross-architecture builds:
     ```bash
     docker run --privileged --rm tonistiigi/binfmt --install all
     ```

3. **Build verification:**
   - Verify dynamic linking with `ldd ./service-user`
   - Check glibc requirement with `readelf -V ./service-user | grep GLIBC`

## References

- goreleaser documentation: https://goreleaser.com/
- Static vs dynamic linking: https://golang.org/cmd/link/
- glibc versioning: https://sourceware.org/glibc/wiki/Versioning
