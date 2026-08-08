#!/usr/bin/env bash
# Single source of truth for CI/code-quality checks.
# Invoked locally by hooks/pre-push (blocking) and by .github/workflows/ci.yml
# (confirmation only). Both run this exact script so local and CI cannot drift.
set -e

echo "ci-check: build..."
go build ./...

echo "ci-check: vet..."
go vet ./...

echo "ci-check: test..."
go test ./...

echo "ci-check passed."
