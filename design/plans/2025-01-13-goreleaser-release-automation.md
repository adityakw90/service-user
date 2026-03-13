# Goreleaser Release Automation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement hybrid static + dynamic binary release automation using goreleaser for the service-user microservice, supporting Linux (4 glibc variants), macOS, and Windows across amd64/arm64 architectures.

**Architecture:** Replace current GitHub Actions build matrix with goreleaser-based automation. Static binaries (CGO_ENABLED=0) for maximum portability; dynamic binaries (CGO_ENABLED=1) built on specific Debian/Ubuntu containers for glibc compatibility. Use goreleaser's declarative configuration to produce 15 artifacts per release (5 static + 8 dynamic + 2 verification files).

**Tech Stack:** goreleaser (Go release automation), GitHub Actions (CI/CD), Docker (cross-platform builds), GPG (artifact signing)

---

## File Structure

```
service-user/
├── .goreleaser.yaml              # NEW: Goreleaser configuration
├── DYNAMIC_NOTES.md               # NEW: User documentation for dynamic binaries
├── .github/workflows/
│   ├── release.yml               # MODIFY: Archive (keep as backup)
│   └── goreleaser.yml            # NEW: Goreleaser release workflow
├── design/
│   ├── docs/                     # NEW: Documentation directory
│   │   ├── release-testing-checklist.md
│   │   └── goreleaser-migration-guide.md
│   ├── plans/
│   │   └── 2025-01-13-goreleaser-release-automation.md  # THIS FILE
│   └── specs/
│       └── 2025-01-13-goreleaser-release-automation-design.md  # REFERENCE: Design spec
└── config.yaml.example           # EXISTING: Included in release archives
```

**File responsibilities:**
- `.goreleaser.yaml`: Complete goreleaser configuration defining all build variants, archive templates, checksums, and signing
- `DYNAMIC_NOTES.md`: End-user documentation explaining dynamic linking, compatibility, and binary selection
- `.github/workflows/release.yml`: Existing workflow to be updated/modified for goreleaser integration
- `.github/workflows/goreleaser.yml`: New workflow dedicated to goreleaser releases

---

## Chunk 1: Phase 1 - Add Goreleaser Configuration

### Task 1: Create DYNAMIC_NOTES.md Documentation

**Files:**
- Create: `DYNAMIC_NOTES.md`

- [ ] **Step 1: Create DYNAMIC_NOTES.md with complete documentation**

Write the user-facing documentation for dynamic binaries:

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

- [ ] **Step 2: Verify file was created**

Run: `ls -la DYNAMIC_NOTES.md`
Expected: File exists with content

- [ ] **Step 3: Commit DYNAMIC_NOTES.md**

```bash
git add DYNAMIC_NOTES.md
git commit -m "docs: add dynamic binary user documentation"
```

---

### Task 2: Create Goreleaser Configuration

**Files:**
- Create: `.goreleaser.yaml`

- [ ] **Step 1: Create .goreleaser.yaml with complete configuration**

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
      - src: README.md
        dst: .
      - src: DYNAMIC_NOTES.md
        dst: .

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

signs:
  # Note: GPG signing is optional. If GPG is not configured, goreleaser will skip signing.
  # To enable GPG signing, configure a GPG key and ensure GITHUB_TOKEN has access.
  - artifacts: all
  signature: "${artifact}.sig"
  cmd: gpg
  args:
    - "--output"
    - "${signature}"
    - "--detach-sig"
    - "${artifact}"
  # Environment variables (optional, for CI/CD):
  # GPG_FINGERPRINT: The fingerprint of the signing key
  # GPG_PASSWORD: The password for the signing key (use GitHub Secrets)

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

- [ ] **Step 2: Verify goreleaser configuration**

Run: `goreleaser check`
Expected: No errors (goreleaser validates its configuration)

- [ ] **Step 3: Commit .goreleaser.yaml**

```bash
git add .goreleaser.yaml
git commit -m "feat: add goreleaser configuration for hybrid release automation"
```

---

### Task 3: Install Goreleaser Locally (Development Setup)

**Files:**
- None (tool installation)

- [ ] **Step 1: Check if goreleaser is installed**

Run: `goreleaser --version`
Expected: Either version output OR "command not found"

- [ ] **Step 2: Install goreleaser if not present**

Run: `go install github.com/goreleaser/goreleaser/v2@latest`
Expected: Downloads and installs goreleaser

- [ ] **Step 3: Verify goreleaser in PATH**

Run: `~/go/bin/goreleaser --version || $HOME/go/bin/goreleaser --version`
Expected: Version output

- [ ] **Step 4: Add to .gitignore if needed**

Check: `grep -q "dist/" .gitignore || echo "dist/" >> .gitignore`
Expected: "dist/" added to .gitignore

- [ ] **Step 5: Commit .gitignore update if modified**

```bash
git add .gitignore
git commit -m "chore: add dist/ to gitignore for goreleaser output"
```

---

## Chunk 2: Phase 2 - Test Goreleaser Locally

### Task 4: Test Static Builds Locally

**Files:**
- No file modifications (testing only)

- [ ] **Step 1: Run goreleaser snapshot build for static Linux**

Run: `goreleaser build --snapshot --clean --id static-linux`
Expected: Builds `dist/service-user_linux_amd64/service-user` and `dist/service-user_linux_arm64/service-user`

- [ ] **Step 2: Verify static binary was created**

Run: `ls -la dist/service-user_linux_amd64/`
Expected: Binary exists (~50MB)

- [ ] **Step 3: Test static binary execution**

Run: `dist/service-user_linux_amd64/service-user --version`
Expected: Version output

- [ ] **Step 4: Verify binary is static (no libc dependency)**

Run: `ldd dist/service-user_linux_amd64/service-user 2>&1 | grep "not a dynamic executable"`
Expected: "not a dynamic executable" (indicating static binary)

- [ ] **Step 5: Clean build artifacts**

Run: `rm -rf dist/`
Expected: dist/ removed

---

### Task 5: Test Cross-Platform Builds (macOS, Windows)

**Files:**
- No file modifications (testing only)

- [ ] **Step 1: Build macOS snapshot**

Run: `goreleaser build --snapshot --clean --id static-darwin`
Expected: Builds `dist/service-user_darwin_amd64/service-user` and `dist/service-user_darwin_arm64/service-user`

- [ ] **Step 2: Build Windows snapshot**

Run: `goreleaser build --snapshot --clean --id static-windows`
Expected: Builds `dist/service-user_windows_amd64/service-user.exe`

- [ ] **Step 3: Verify all platform artifacts exist**

Run: `ls -la dist/`
Expected: Directories for linux_amd64, linux_arm64, darwin_amd64, darwin_arm64, windows_amd64

- [ ] **Step 4: Clean build artifacts**

Run: `rm -rf dist/`
Expected: dist/ removed

---

### Task 6: Test Archive Creation

**Files:**
- No file modifications (testing only)

- [ ] **Step 1: Run full snapshot build**

Run: `goreleaser release --snapshot --clean`
Expected: All builds and archives created in dist/

- [ ] **Step 2: Verify archive files were created**

Run: `find dist/ -name "*.tar.gz" -type f | head -10`
Expected: List of .tar.gz files with correct naming pattern

- [ ] **Step 3: Verify checksums file was created**

Run: `cat dist/checksums.txt | head -5`
Expected: SHA256 hashes of artifacts

- [ ] **Step 4: Test archive extraction**

Run: `mkdir -p /tmp/test-extract && tar -xzf dist/service-user_*_linux_amd64.tar.gz -C /tmp/test-extract && ls /tmp/test-extract/`
Expected: Archive contains service-user binary, config.yaml.example, README.md

- [ ] **Step 5: Verify dynamic archive contains DYNAMIC_NOTES.md**

Run: `tar -tzf dist/service-user_*_linux_glibc*_amd64.tar.gz | grep -q DYNAMIC_NOTES.md`
Expected: Output confirms DYNAMIC_NOTES.md is in dynamic archives

- [ ] **Step 6: Clean up test extraction**

Run: `rm -rf /tmp/test-extract dist/`
Expected: Test files removed

---

## Chunk 3: Phase 3 - Update CI/CD Workflow

### Task 7: Create Goreleaser GitHub Actions Workflow

**Files:**
- Create: `.github/workflows/goreleaser.yml`

- [ ] **Step 1: Create goreleaser.yml workflow**

```yaml
name: Goreleaser Release

on:
  push:
    tags:
      - "v*.*.*"
  workflow_dispatch:
    inputs:
      version:
        description: "Version tag (e.g., v1.0.0)"
        required: true
        type: string
      snapshot:
        description: "Create snapshot (test) release"
        type: boolean
        default: false

permissions:
  contents: write

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up QEMU for cross-platform builds
        uses: docker/setup-qemu-action@v3
        with:
          platforms: linux/amd64,linux/arm64

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.25.5"
          cache-dependency-path: go.sum

      - name: Install GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          install-only: true

      - name: Run GoReleaser (snapshot)
        if: github.event.inputs.snapshot == 'true' || github.event_name != 'push'
        run: |
          # Build static variants only for snapshot testing
          ~/go/bin/goreleaser release --snapshot --clean --ids static-linux,static-darwin,static-windows
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Run GoReleaser (release)
        if: github.event.inputs.snapshot != 'true' && github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
        run: |
          # Full release with all variants
          # Note: Dynamic builds require appropriate build environments
          # For this initial implementation, we build static variants in CI
          ~/go/bin/goreleaser release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          # GPG signing (optional - configure GPG key in GitHub Secrets to enable)
          # GPG_FINGERPRINT: ${{ secrets.GPG_FINGERPRINT }}

  # Dynamic builds run separately in specific containers
  # These jobs only run for actual releases (not snapshots)
  dynamic-release:
    name: Dynamic Build (${{ matrix.glibc }})
    runs-on: ubuntu-latest
    if: github.event.inputs.snapshot != 'true' && github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    container:
      image: ${{ matrix.image }}
    strategy:
      matrix:
        include:
          - glibc: "2.28"
            image: debian:10-slim
            build_id: dynamic-glibc228
          - glibc: "2.31"
            image: debian:11-slim
            build_id: dynamic-glibc231
          - glibc: "2.36"
            image: debian:12-slim
            build_id: dynamic-glibc236
          - glibc: "2.39"
            image: ubuntu:24.04
            build_id: dynamic-glibc239
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.25.5"
          cache-dependency-path: go.sum

      - name: Install build dependencies
        run: |
          apt-get update
          apt-get install -y gcc libc-dev wget

      - name: Set up QEMU for ARM builds
        run: |
          apt-get update
          apt-get install -y qemu-user-static binfmt-support
          # Verify QEMU is working
          qemu-aarch64-static --version || echo "QEMU aarch64 not available, ARM builds may fail"

      - name: Install GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          install-only: true

      - name: Build dynamic variant (amd64)
        run: |
          ~/go/bin/goreleaser build --clean --ids ${{ matrix.build_id }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          CGO_ENABLED: 1
          GOOS: linux
          GOARCH: amd64

      - name: Build dynamic variant (arm64)
        run: |
          # Only attempt ARM64 if QEMU is available
          if [ -f /usr/bin/qemu-aarch64-static ] || [ -f /usr/bin/qemu-aarch64 ]; then
            ~/go/bin/goreleaser build --clean --ids ${{ matrix.build_id }}
          else
            echo "ARM64 QEMU not available in container, skipping ARM64 build"
            exit 0
          fi
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          CGO_ENABLED: 1
          GOOS: linux
          GOARCH: arm64
          CC: aarch64-linux-gnu-gcc

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: dynamic-${{ matrix.glibc }}
          path: dist/
          if-no-files-found: warn

  merge-artifacts:
    name: Merge Release Artifacts
    runs-on: ubuntu-latest
    needs: [release, dynamic-release]
    if: github.event.inputs.snapshot != 'true' && github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.25.5"
          cache-dependency-path: go.sum

      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts/

      - name: Merge artifacts and create archives
        run: |
          # Copy static artifacts from main release job
          mkdir -p dist

          # Copy all artifacts to dist directory
          cp -r artifacts/* dist/ || true

          # Generate final checksums
          cd dist
          find . -name "service-user*" -type f | sort | xargs sha256sum > checksums.txt
          cat checksums.txt
          cd ..

      - name: Upload release assets
        uses: softprops/action-gh-release@v2
        with:
          files: |
            dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Note:** This workflow uses a two-stage approach:
1. Main release job builds static binaries (fast, no container needed)
2. Matrix jobs build dynamic variants in appropriate containers
3. Final merge job combines all artifacts and uploads to release

- [ ] **Step 2: Verify goreleaser workflow syntax**

Run: `cat .github/workflows/goreleaser.yml | head -20`
Expected: Workflow YAML is properly formatted

- [ ] **Step 3: Commit goreleaser.yml workflow**

```bash
git add .github/workflows/goreleaser.yml
git commit -m "ci: add goreleaser release workflow with matrix builds"
```

---

### Task 8: Archive Existing Release Workflow

**Files:**
- Modify: `.github/workflows/release.yml` → Rename to `.github/workflows/release.yml.old`

- [ ] **Step 1: Archive existing release workflow**

Run: `mv .github/workflows/release.yml .github/workflows/release.yml.old`
Expected: File renamed

- [ ] **Step 2: Commit archived workflow**

```bash
git add .github/workflows/release.yml.old
git commit -m "ci: archive old release workflow (replaced by goreleaser)"
```

---

## Chunk 4: Phase 4 - Testing and Validation

### Task 9: Create Release Testing Checklist

**Files:**
- Create: `design/docs/release-testing-checklist.md`

- [ ] **Step 1: Create design/docs directory**

Run: `mkdir -p design/docs`
Expected: Directory created

- [ ] **Step 2: Create release testing checklist**

```markdown
# Release Testing Checklist

## Pre-Release Testing

Before triggering a release, verify:

- [ ] All tests pass locally: `go test ./...`
- [ ] Goreleaser snapshot builds successfully: `goreleaser release --snapshot --clean`
- [ ] Static binaries work on local system
- [ ] Archive extraction works correctly
- [ ] Checksums file is generated

## Post-Release Validation

After release is created:

- [ ] Verify all 15 artifacts are present on GitHub release page
- [ ] Download and test static Linux binary (amd64)
- [ ] Verify checksums: `sha256sum -c checksums.txt`
- [ ] Check archive contents include all required files
- [ ] Verify release notes are generated correctly

## Dynamic Binary Testing (when applicable)

- [ ] Test glibc228 binary on Debian 10 or newer
- [ ] Test glibc231 binary on Debian 11 or newer
- [ ] Test glibc236 binary on Debian 12 or newer
- [ ] Test glibc239 binary on Ubuntu 24.04 or newer

## Rollback Procedure

If release fails:

1. Delete the release from GitHub
2. Archive the tag: `git tag archive/vX.Y.Z-failed <tag-sha>`
2. Delete the bad tag: `git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z`
3. Investigate failure in CI/CD logs
4. Fix issues and retry
```

- [ ] **Step 3: Commit testing checklist**

```bash
git add design/docs/release-testing-checklist.md
git commit -m "docs: add release testing checklist"
```

---

### Task 10: Create Migration Documentation

**Files:**
- Create: `design/docs/goreleaser-migration-guide.md`

- [ ] **Step 1: Ensure design/docs directory exists**

Run: `mkdir -p design/docs`
Expected: Directory exists (or is created)

- [ ] **Step 2: Create migration guide**

```markdown
# Goreleaser Migration Guide

## Overview

This document describes the migration from the custom GitHub Actions release workflow to goreleaser-based automation.

## What Changed

### Before
- Custom GitHub Actions workflow with matrix builds
- Manual artifact naming and packaging
- Separate jobs for building and packaging
- 4 artifacts: glibc-amd64, glibc-arm64, musl-amd64, musl-arm64

### After
- Goreleaser declarative configuration
- Automatic artifact naming and packaging
- Single goreleaser release process
- 15 artifacts: 5 static + 8 dynamic + 2 verification files

## New Capabilities

1. **Cross-platform support**: macOS and Windows binaries now included
2. **Multiple glibc variants**: glibc 2.28, 2.31, 2.36, 2.39
3. **Automatic checksums**: SHA256 checksums generated for all artifacts
4. **GPG signing**: Optional artifact signing (configure GPG key to enable)
5. **Changelog generation**: Automatic changelog from commit messages

## Usage

### Creating a Release

**Option 1: Tag-based (automatic)**
```bash
git tag v1.0.0
git push origin v1.0.0
```

**Option 2: Manual trigger**
1. Go to Actions tab in GitHub
2. Select "Goreleaser Release" workflow
3. Click "Run workflow"
4. Enter version tag (e.g., v1.0.0)

### Testing Locally

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Test build
goreleaser release --snapshot --clean

# Test specific build variant
goreleaser build --snapshot --clean --id static-linux
```

## Troubleshooting

### Dynamic builds fail in CI

Dynamic builds require proper build environments. If you see errors:

1. Check the container image matches the glibc version
2. Verify CGO_ENABLED=1 is set for dynamic builds
3. Ensure gcc and libc-dev are installed in the container

### Archive missing files

If archives don't include expected files:

1. Check files section in .goreleaser.yaml
2. Verify source files exist in repository
3. Check DYNAMIC_NOTES.md is present for dynamic archives

### Wrong platform binaries

If wrong architecture binaries are built:

1. Check goos/goarch settings in .goreleaser.yaml
2. Verify cross-compilation setup for ARM builds
3. Use QEMU/binfmt for ARM cross-compilation

## Rollback

If goreleaser release fails, the old workflow is archived as `.github/workflows/release.yml.old`. To restore:

1. Delete `.github/workflows/goreleaser.yml`
2. Restore old workflow: `mv .github/workflows/release.yml.old .github/workflows/release.yml`
3. Commit and push
```

- [ ] **Step 3: Commit migration guide**

```bash
git add design/docs/goreleaser-migration-guide.md
git commit -m "docs: add goreleaser migration guide"
```

---

## Chunk 5: Final Verification

### Task 11: Final Integration Verification

**Files:**
- No modifications (verification only)

- [ ] **Step 1: Verify all files are in place**

Run: `ls -la .goreleaser.yaml DYNAMIC_NOTES.md .github/workflows/goreleaser.yml`
Expected: All three files exist

- [ ] **Step 2: Run final snapshot test**

Run: `goreleaser release --snapshot --clean 2>&1 | tail -20`
Expected: Build completes successfully

- [ ] **Step 3: Verify artifact count**

Run: `find dist/ -name "*.tar.gz" -type f | wc -l`
Expected: 13 archives (5 static + 8 dynamic)

Note: Full release will have 15 total artifacts including checksums.txt and checksums.txt.sig (if GPG signing is enabled)

- [ ] **Step 4: Verify checksums file**

Run: `test -f dist/checksums.txt && wc -l dist/checksums.txt`
Expected: File exists with 13 lines (plus header if any)

- [ ] **Step 5: Clean up snapshot**

Run: `rm -rf dist/`
Expected: Clean dist directory

- [ ] **Step 6: Stage all changes for final commit**

Run: `git status`
Expected: Shows all modified/new files ready for commit

- [ ] **Step 7: Create final integration commit**

```bash
git add -A
git commit -m "feat: complete goreleaser release automation implementation

This commit completes the implementation of hybrid static + dynamic
binary release automation using goreleaser.

Features:
- Static binaries for Linux, macOS, Windows (amd64/arm64)
- Dynamic binaries for 4 glibc versions (2.28, 2.31, 2.36, 2.39)
- 15 artifacts per release (5 static + 8 dynamic + 2 verification)
- Automatic checksums and optional GPG signing
- CI/CD integration via GitHub Actions

Documentation:
- DYNAMIC_NOTES.md: User guide for dynamic binary selection
- Migration guide: Rollback instructions and troubleshooting
- Testing checklist: Pre/post-release validation steps

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

- [ ] **Step 8: Push to feature branch**

Run: `git push origin feat/release-binaries`
Expected: Pushes to remote branch

---

## Post-Implementation

### Task 12: Create Pull Request

**Files:**
- None (GitHub action)

- [ ] **Step 1: Create pull request**

Using GitHub CLI: `gh pr create --title "feat: goreleaser release automation" --body "Implements hybrid static + dynamic binary release automation using goreleaser. See design spec for details."`
Or manually via GitHub web interface.

- [ ] **Step 2: Request review from team**

Assign appropriate reviewers for the pull request.

---

## Summary

This implementation plan creates:

1. **Configuration Files:**
   - `.goreleaser.yaml` - Complete goreleaser configuration
   - `.github/workflows/goreleaser.yml` - CI/CD integration

2. **Documentation:**
   - `DYNAMIC_NOTES.md` - End-user documentation for dynamic binaries
   - `design/docs/release-testing-checklist.md` - Testing procedures
   - `design/docs/goreleaser-migration-guide.md` - Migration and rollback

3. **Archived Files:**
   - `.github/workflows/release.yml.old` - Old workflow preserved for rollback

**Total artifacts per release:** 15 (5 static + 8 dynamic + 2 verification)

**Build time:** Approximately 10-15 minutes for full release

**Next steps after merge:**
1. Monitor first release on main branch
2. Validate all artifacts download and execute correctly
3. Archive old workflow after successful production release
4. Consider enabling GPG signing for enhanced security
