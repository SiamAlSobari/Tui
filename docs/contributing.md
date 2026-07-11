# Contributing to dbbee 🐝

Thank you for your interest in contributing to `dbbee`! We welcome community contributions, bug reports, and suggestions.

---

## 🛠️ Development Setup

To start developing on `dbbee`, follow these steps:

1. **Prerequisites**: Make sure you have Go 1.21+ installed on your computer.
2. **Clone the Repository**:
   ```bash
   git clone https://github.com/SiamAlSobari/dbbee.git
   cd dbbee
   ```
3. **Run Code Locally**:
   You can run `dbbee` immediately against a sample database:
   ```bash
   go run ./cmd/dbbee/ data/sample.db
   ```
4. **Compile Binaries**:
   ```bash
   go build -o bin/dbbee ./cmd/dbbee
   ```

---

## 🧪 Testing Guidelines

We enforce high test coverage. Before submitting any pull requests, ensure that all tests compile and pass.

- **Run all unit tests**:
  ```bash
  go test -v ./...
  ```
- **Run benchmarks**:
  We benchmark startup operations to guarantee performance stays within optimal limits (<50ms):
  ```bash
  go test -bench=. ./internal/db
  ```

---

## 📐 Coding Conventions

- **CGO-Free Constraint**: `dbbee` must remain strictly CGO-free. Do not import package dependencies that rely on C libraries.
- **Go Style**: Run `go fmt ./...` and `go vet ./...` to keep code formatted and clean.
- **Comments**: Keep comments clear, descriptive, and maintain documentation updates.
- **Elm MVU Architecture**: When modifying the TUI, follow the Bubble Tea Model-View-Update design pattern. Components should be decoupled and communicate via Bubble Tea messages (`tea.Msg`).
