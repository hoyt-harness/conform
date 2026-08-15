#!/usr/bin/env bash
# SPDX-License-Identifier: GPL-3.0-or-later
# Single source of truth for CI/code-quality checks.
# Invoked locally by hooks/pre-push (blocking) and by .github/workflows/ci.yml
# (confirmation only). Both run this exact script so local and CI cannot drift.
set -e

echo "ci-check: build..."
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
go build -ldflags "-X main.version=${VERSION}" ./...

echo "ci-check: vet..."
go vet ./...

echo "ci-check: test..."
go test ./...

echo "ci-check passed."
