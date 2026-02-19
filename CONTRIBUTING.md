# Contributing to enviar

Thank you for your interest in contributing! This document provides guidelines
and information on how to get involved.

## Prerequisites

- **Go 1.24+**
- **Redis** — a running instance for integration testing
- **Git**

## Getting Started

1. Fork the repository on GitHub.
2. Clone your fork:

   ```bash
   git clone https://github.com/<your-username>/enviar.git
   cd enviar
   ```

3. Ensure dependencies are up to date:

   ```bash
   go mod tidy
   ```

4. Run the tests:

   ```bash
   go test -v ./...
   ```

## Development Workflow

### Branching

- Create a feature branch from `main`:

  ```bash
  git checkout -b feature/my-change
  ```

- Use descriptive branch names: `fix/nil-body-panic`, `feature/batch-enqueue`,
  `docs/improve-readme`, etc.

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`).
- Run `go vet ./...` before committing.
- Keep exported identifiers well-documented with GoDoc comments.
- Prefer table-driven tests for comprehensive coverage.
- Avoid unnecessary dependencies — this package should stay lightweight.

### Testing

- All new functionality must include tests.
- Tests should not require a live Redis connection unless tagged with
  `//go:build integration`. Standard tests must be self-contained.
- Run the full suite before submitting:

  ```bash
  go test -v -race -count=1 ./...
  ```

### Commit Messages

Write clear, concise commit messages:

```
Add batch enqueue method

Introduce EnqueueBatch for enqueuing multiple jobs in a single
Redis pipeline. Includes unit tests and documentation.
```

- Use the imperative mood ("Add", not "Added" or "Adds").
- Keep the first line under 72 characters.
- Add a blank line before the body if more detail is needed.

## Submitting Changes

1. Push your branch to your fork:

   ```bash
   git push origin feature/my-change
   ```

2. Open a Pull Request against `main` on the upstream repository.
3. Fill in the PR template with:
   - A summary of changes.
   - How to test the changes.
   - Any related issues.

## Reporting Issues

- Use the GitHub issue tracker.
- Include steps to reproduce, expected behavior, and actual behavior.
- Include your Go version (`go version`) and OS.

## Code of Conduct

Be respectful and constructive. We follow the
[Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE).
