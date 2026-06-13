# Contributing to Loadster

First off, thank you for considering contributing to Loadster! It's people like you that make Loadster such a great tool for the community.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct. Please treat all contributors with respect.

## How Can I Contribute?

### Reporting Bugs

If you find a bug, please open an issue and include:
* Your operating system and Go version.
* A clear and descriptive title.
* Steps to reproduce the behavior.
* Expected vs actual behavior.
* Any relevant logs or screenshots.

### Suggesting Enhancements

We welcome new feature suggestions! When opening an enhancement issue, please describe:
* The problem you are trying to solve.
* Your proposed solution or feature.
* Any alternative solutions you've considered.

### Local Development Setup

To get your environment ready for development:

1. **Fork and Clone:**
   Fork the repository on GitHub and clone it locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/Loadster.git
   cd Loadster
   ```

2. **Ensure Go is Installed:**
   Loadster requires Go 1.24 or higher.

3. **Install Dependencies:**
   ```bash
   go mod download
   ```

4. **Run the Code:**
   You can run Loadster directly from the source:
   ```bash
   go run main.go init
   go run main.go run --config test_scenario.yaml
   ```

### Running Tests

Before submitting a pull request, ensure all tests pass:

```bash
# Run unit and integration tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.txt ./...
```

### Linting

We enforce strict linting using `golangci-lint`. Please ensure your code passes the linter:

```bash
# Install golangci-lint (if not installed)
# Run the linter
golangci-lint run
```

## Pull Request Process

1. Create a new branch for your feature or bugfix (`git checkout -b feature/my-feature`).
2. Make your changes and write tests if applicable.
3. Ensure the code is properly formatted (`go fmt ./...`).
4. Ensure the linter and tests pass.
5. Commit your changes with descriptive commit messages.
6. Push your branch to your fork.
7. Open a Pull Request against the `main` branch.

Once your PR is opened, a maintainer will review your code. We may request changes, but we will always aim to be helpful and constructive!

Thank you for contributing!
