# Contributing to StreamGoGambler

Thank you for your interest in contributing to StreamGoGambler! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful and constructive in all interactions. We welcome contributors of all experience levels.

## Getting Started

### Prerequisites

- Go 1.25 or later
- golangci-lint for linting
- pre-commit (optional but recommended)

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/streamgogambler.git
   cd streamgogambler
   ```
3. Install development tools:
   ```bash
   make install-tools
   ```
4. Install pre-commit hooks (optional):
   ```bash
   pip install pre-commit
   make pre-commit-install
   ```

## Development Workflow

### Before Making Changes

1. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make sure tests pass:
   ```bash
   make test
   ```

### Making Changes

1. Write your code following the existing style
2. Add tests for new functionality
3. Run the linter:
   ```bash
   make lint
   ```
4. Run tests:
   ```bash
   make test
   ```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `make lint` before committing
- Keep functions focused and small
- Add comments for exported functions
- Use meaningful variable names

### Testing

- Write unit tests for new functions
- Add integration tests for complex interactions
- Run fuzz tests for parser changes:
  ```bash
  make fuzz
  ```
- Test with race detector when modifying concurrent code:
  ```bash
  make test-race
  ```

### Commit Messages

Use clear, descriptive commit messages:

```
Add feature X

- Implemented Y
- Updated Z
- Added tests for A
```

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Run the full check:
   ```bash
   make check
   ```
5. Submit a pull request with:
   - Clear description of changes
   - Link to related issues (if any)
   - Screenshots for UI changes (if applicable)

### PR Checklist

- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] Linter passes (`make lint`)
- [ ] All tests pass (`make test`)
- [ ] No race conditions (`make test-race`)

## Reporting Issues

### Bug Reports

Include:
- Go version (`go version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs (with sensitive info redacted)

### Feature Requests

Include:
- Use case description
- Proposed solution (if any)
- Alternative approaches considered

## Questions?

Open an issue with the "question" label for any questions about contributing.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
