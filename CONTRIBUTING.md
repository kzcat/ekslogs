# Contributing to ekslogs

Thank you for your interest in contributing to ekslogs! This document provides guidelines for contributing to the project.

## Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - Make
   - Git

2. **Clone the repository**
   ```bash
   git clone https://github.com/kzcat/ekslogs.git
   cd ekslogs
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Build the project**
   ```bash
   make build
   ```

5. **Run tests**
   ```bash
   make test
   ```

## Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write code following Go best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   make test
   make test-coverage
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

## Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused

## Testing

- Write unit tests for new functions
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Test edge cases and error conditions

## Commit Message Format

Use conventional commits format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test changes
- `refactor:` for code refactoring

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add a clear description of changes
4. Link any related issues
5. Request review from maintainers

## Reporting Issues

When reporting issues, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Error messages or logs

## Questions?

Feel free to open an issue for questions or discussions about the project.