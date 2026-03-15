# bin/bash

git tag -d v0.0.0-goreleaser 2>/dev/null || true
git tag v0.0.0-goreleaser
git push origin v0.0.0-goreleaser --force
