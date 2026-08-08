# Contributing

Contributions are welcome. Please follow these guidelines.

## How to Contribute

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/your-feature`).
3. Make your changes and commit them with clear conventional commit messages.
4. Ensure your code passes `hooks/ci-check.sh` locally.
5. Submit a pull request.

## Coding Standards

This project follows the
[Positronikal Coding Standards](https://github.com/Positronikal/PositronikalCodingStandards).
Key points for Go code:

- `gofmt`-formatted (enforced by `go vet` in CI)
- No external dependencies — standard library only
- Test coverage for all new dispatch paths and file type detections
- One status line per file: the output contract in `contracts/cli.md` is stable

## Running Tests

```bash
go test -race ./...
```

## Legal

By contributing, you agree that your contributions are licensed under the
GNU General Public License v3.0, the same license as this project.
