# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

tdl is a Telegram Downloader CLI tool written in Go that provides file downloads, uploads, forwarding, and message/member export functionality. It uses the gotd/td library for Telegram MTProto API communication.

## Build & Development Commands

### Building
```bash
# Build for current platform using goreleaser
make build
# Output will be in .tdl/dist/

# Build directly with Go
go build

# Build all packages for release (multi-platform)
make packaging
```

### Testing
```bash
# Run unit tests (excludes e2e tests)
go test -v $(go list ./... | grep -v /test)

# Run e2e tests (requires Ginkgo and Teamgram server)
go install github.com/onsi/ginkgo/v2/ginkgo
ginkgo -v -r ./test
```

### Linting
```bash
# Lint with golangci-lint
golangci-lint run

# Lint specific modules (use for core/extension)
cd core && golangci-lint run
cd extension && golangci-lint run

# Auto-format code (gci and gofumpt)
golangci-lint run --fix
```

### Running Tests for Single Package
```bash
# Test specific package
go test -v ./pkg/kv
go test -v ./app/dl

# Test with coverage
go test -v -cover ./pkg/...
```

## Architecture

### Module Structure
This is a Go workspace project with three modules:
- **Root module** (`github.com/iyear/tdl`): CLI layer, commands, and high-level app logic
- **core** (`github.com/iyear/tdl/core`): Core business logic (downloader, uploader, forwarder, tclient, storage)
- **extension** (`github.com/iyear/tdl/extension`): Extension framework for user plugins

### Key Directories
- `cmd/`: Cobra command definitions (root, dl, up, forward, chat, login, etc.)
- `app/`: Command implementation layer (dl, forward, login, upload, chat, etc.)
- `core/`: Core business logic modules
  - `downloader/`: Download orchestration
  - `uploader/`: Upload orchestration
  - `forwarder/`: Message forwarding logic
  - `tclient/`: Telegram client wrapper
  - `dcpool/`: Telegram DC (data center) connection pooling
  - `storage/`: Storage abstraction interface
  - `tmedia/`: Media processing (thumbnails, video conversion)
- `pkg/`: Shared utilities and packages
  - `kv/`: Key-value storage drivers (legacy, bolt, file)
  - `tclient/`: Telegram client construction utilities
  - `extensions/`: Extension manager
  - `texpr/`: Template expression evaluation
  - `tpath/`: Telegram file path utilities
- `test/`: E2E tests using Ginkgo
- `extension/`: Extension SDK and interfaces

### Command Flow
1. User invokes CLI command (e.g., `tdl dl`)
2. `cmd/` package parses flags and validates input
3. `cmd/` calls into `app/` layer for command implementation
4. `app/` layer orchestrates `core/` modules to execute business logic
5. `core/` modules interact with Telegram API via `tclient`

### Storage System
tdl uses a pluggable key-value storage system for session data and metadata:
- **Bolt** (default since v0.14.0): BoltDB-backed storage at `~/.tdl/data`
- **Legacy**: Custom KV format at `~/.tdl/data.kv` (deprecated)
- **File**: Single-file storage for testing
- Storage is namespace-aware (use `-n` flag to switch namespaces)

### Telegram Client Architecture
- `pkg/tclient` constructs authenticated Telegram clients
- `core/tclient` provides `RunWithAuth()` for ensuring authentication
- `core/dcpool` manages multiple DC connections for parallel transfers
- Sessions are persisted in the KV storage system

### Extension System
- Extensions are executable binaries in `~/.tdl/extensions/`
- Extensions receive JSON input via stdin and return JSON via stdout
- Extensions are dynamically loaded as cobra commands
- Extension manager handles installation from GitHub releases or local files

## Code Style & Conventions

### Import Grouping (gci)
Imports must be organized in this order:
1. Standard library
2. External dependencies
3. `github.com/iyear/tdl` packages
4. Dot imports

### Error Handling
Use `github.com/go-faster/errors` for error wrapping:
```go
return errors.Wrap(err, "descriptive context")
```

### Logging
Use `go.uber.org/zap` via `core/logctx`:
```go
logctx.From(ctx).Info("message", zap.String("key", value))
```

### Context Passing
- Always pass `context.Context` as first parameter
- Use context for cancellation, logging, and storage access
- Store logger in context via `logctx.With()`
- Store KV storage in context via `kv.With()`

## Testing Strategy

### Unit Tests
- Place unit tests alongside source files (`*_test.go`)
- Use standard `testing` package for unit tests
- Mock external dependencies when needed

### E2E Tests
- Located in `test/` directory
- Use Ginkgo/Gomega framework
- Run against local Teamgram server (Telegram-compatible test server)
- Test complete workflows (login, download, upload, forward, chat operations)

## Common Patterns

### Creating Telegram Client
```go
o, err := tOptions(ctx)
client, err := tclient.New(ctx, o, false, middlewares...)
err := tclientcore.RunWithAuth(ctx, client, func(ctx context.Context) error {
    // Use authenticated client
})
```

### Accessing KV Storage
```go
stg := kv.From(ctx)
kvd, err := stg.Open("namespace")
defer kvd.Close()
```

### Progress Reporting
Use `pkg/prog` for progress tracking with multiple concurrent items.

## Dependencies

### Core Telegram Library
- `github.com/gotd/td`: Low-level MTProto implementation
- `github.com/gotd/contrib`: Additional Telegram utilities

### CLI Framework
- `github.com/spf13/cobra`: Command structure
- `github.com/spf13/viper`: Configuration management
- `github.com/ivanpirog/coloredcobra`: Colored help output

### Storage
- `go.etcd.io/bbolt`: BoltDB for persistent storage

## Version Management

Version information is embedded at build time via goreleaser. See `.goreleaser.yaml` for release configuration.