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
