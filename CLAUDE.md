# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the Audius Protocol Explorer, a Go-based blockchain explorer for the Audius network. It's a web application that indexes and displays blockchain data including blocks, transactions, validators, and content metadata.

## Development Commands

### Running the Application

```bash
# Start all watchers (Tailwind CSS, Templ templates, Go server)
make run

# Or run components individually:
make tailwind-watch  # Watch and compile Tailwind CSS
make templ-watch     # Watch and compile Templ templates
make go-watch        # Watch and run Go server with hot reload
```

The server runs on `localhost:3000`.

### Database

```bash
# Start PostgreSQL in Docker
make pg

# Generate database code from SQL (after modifying db/queries/*.sql or db/migrations/*.sql)
make sqlc

# Watch for SQL changes and auto-regenerate
make sqlc-watch
```

Database configuration:
- Port: 5444 (not standard 5432)
- Database: audiusd
- Password: postgres

### Dependencies

```bash
# Install Tailwind CSS binary (macOS ARM64)
make tailwind-install
```

Required tools:
- `templ` - Template engine for Go (generates `*_templ.go` files)
- `wgo` - File watcher for Go
- `watchexec` - File watcher for sqlc
- `sqlc` - SQL code generator

## Architecture

### Core Components

**Server** (`server/`)
- Echo-based HTTP server with HTMX and SSE endpoints
- Routes defined in `server/server.go:83-117`
- Integrates with audiusd SDK for blockchain data via `trustedNode`
- Refreshes trusted block height every 10 seconds

**Database Layer** (`db/`)
- Uses sqlc for type-safe SQL query generation
- Migrations in `db/migrations/`
- Queries in `db/queries/` (reads.sql, writes.sql)
- Generated code: `db/*.sql.go` (DO NOT EDIT directly)

**Templates** (`templates/`)
- Uses `a-h/templ` for type-safe HTML templates
- Generates `*_templ.go` files (DO NOT EDIT directly)
- Layouts in `templates/layouts/`
- Pages in `templates/pages/`

**Indexers** (`indexers/`)
- Background processes that index blockchain data into the database
- Implement the `Indexer` interface from `indexers/indexer.go`
- Currently placeholder in `main.go:24-27`

**Entity Models** (`em/`)
- Domain models for Audius entities (users, tracks, playlists, etc.)
- Not actively used in current implementation

**Config** (`config/`)
- Application configuration struct
- Currently returns empty values - needs environment variable integration

**Assets** (`assets/`)
- Embedded static files (CSS, JS, images)
- Tailwind CSS input: `assets/input.css`
- Tailwind CSS output: `assets/css/output.css`

### Key Patterns

1. **Database Access**: All database operations use sqlc-generated `*Queries` methods
2. **Template Compilation**: Templ templates must be compiled before running the Go server
3. **Static Assets**: Assets are embedded in the binary via Go embed
4. **Hot Reload**: `wgo` watches `.go`, `.css`, and `.js` files

### Important Files

- `main.go` - Application entry point with server and indexer goroutines
- `server/server.go` - HTTP server initialization and route definitions
- `sqlc.yaml` - sqlc configuration for database code generation
- `Makefile` - Development workflow automation
