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
3. Delete the bad tag: `git tag -d vX.Y.Z && git push origin :refs/tags/vX.Y.Z`
4. Investigate failure in CI/CD logs
5. Fix issues and retry
