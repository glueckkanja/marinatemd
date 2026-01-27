# Contributing to MarinateMD

Thank you for considering contributing to MarinateMD! This document outlines the process for contributing to this project.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs. actual behavior
- **Go version** and operating system
- **Code samples** or test cases if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear title and description**
- **Use cases** for the enhancement
- **Expected behavior** and rationale
- **Alternative solutions** you've considered

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the coding standards below
3. **Add tests** for any new functionality
4. **Ensure all tests pass** (`go test ./...`)
5. **Run linting** (`golangci-lint run`)
6. **Format your code** (`gofmt -s -w .`)
7. **Commit your changes** using clear commit messages
8. **Push to your fork** and submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- golangci-lint (for linting)
- Make (optional, for running Makefile targets)

### Setup Steps

```bash
# Clone the repository
git clone https://github.com/glueckkanja/marinatemd.git
cd marinatemd

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run

# Build the project
go build -o marinatemd .
```

## Coding Standards

### Go Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (enforced in CI)
- Write clear, idiomatic Go code

### Code Organization

- Keep packages focused and single-purpose
- Place internal packages under `internal/`
- Keep functions small and testable
- Use meaningful variable and function names

### Testing

- Write unit tests for all new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

### Documentation

- Document all exported functions, types, and packages
- Use clear and concise comments
- Update README.md for user-facing changes
- Include examples where helpful

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```text
<type>(<scope>): <subject>

<body>

<footer>
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

Example:

```text
feat(parser): add support for nested object types

Implement parsing logic for deeply nested Terraform object types
to improve schema extraction accuracy.

Closes #42
```

## Review Process

1. Maintainers will review your pull request
2. Feedback will be provided for improvements
3. Once approved, the PR will be merged
4. Your contribution will be included in the next release

## Project Structure

```text
marinatemd/
├── cmd/marinatemd/     # CLI commands
├── internal/
│   ├── config/         # Configuration management
│   ├── hclparse/       # HCL parsing logic
│   ├── schema/         # Schema modeling
│   ├── yamlio/         # YAML I/O operations
│   └── markdown/       # Markdown generation
├── examples/           # Example files
└── docs/               # Documentation
```

## Questions?

Feel free to open an issue with the `question` label if you need clarification on anything.

## Recognition

Contributors will be recognized in release notes and the project README.

Thank you for contributing to MarinateMD! 🎉
