# CoolDB

[![Test](https://github.com/rushikeshg25/cool-db/actions/workflows/test.yml/badge.svg)](https://github.com/rushikeshg25/cool-db/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/badge/Go-1.23.5-blue.svg)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-16.3.3-black.svg)](https://nextjs.org/)

CoolDB is a compact SQL database implemented in Go. It combines a persistent database engine, a gRPC server, an interactive shell, a scriptable command-line client, and a local Next.js query studio in one repository.

The project is currently at version 0.1 and is intended for learning, experimentation, and local demonstrations. It is not ready for production workloads.

## Contents

- [Features](#features)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Quick start](#quick-start)
- [Command-line reference](#command-line-reference)
- [Supported SQL](#supported-sql)
- [Persistence model](#persistence-model)
- [Local Studio](#local-studio)
- [Development](#development)
- [Current limitations](#current-limitations)
- [Roadmap](#roadmap)
- [License](#license)

## Features

- A focused SQL subset with `CREATE TABLE`, `DROP TABLE`, `INSERT`, `SELECT`, `UPDATE`, and `DELETE`.
- Typed columns with integer, text, Boolean, and floating-point values.
- Schema constraints including `PRIMARY KEY`, `UNIQUE`, `NOT NULL`, and `VARCHAR(n)` length validation.
- Persistent database files written through atomic snapshot replacement.
- A gRPC transport shared by the interactive shell and scriptable client.
- One `cool` binary for running the server, opening a shell, and executing queries.
- A local-only query Studio with three interface prototypes.
- Automated Go tests, race detection, static analysis, UI linting, production builds, dependency audits, and security scans.

## Architecture

CoolDB keeps query execution in one database engine and exposes it through two interfaces:

```text
cool shell / cool exec
          |
          v
        gRPC
          |
          v
    Core adapter ---------> Database engine ---------> .cooldb snapshot
                               ^
                               |
Browser -> Next.js routes -> Local HTTP bridge
```

The gRPC path is the primary client interface. The HTTP bridge exists only for the local Studio and is disabled unless an HTTP port is explicitly configured.

| Area | Responsibility |
| --- | --- |
| `cmd/` | Cobra commands for the server, shell, one-shot execution, version output, and completion. |
| `internal/client/` | gRPC client connection and query execution. |
| `internal/core/` | Adapter between the gRPC service and the database engine. |
| `internal/database/` | SQL parsing, type checking, constraints, execution, result formatting, and persistence. |
| `internal/httpapi/` | Restricted local HTTP API used by the Studio. |
| `internal/shell/` | Interactive input and terminal result rendering. |
| `server/` | Server lifecycle, configuration, signal handling, and transport coordination. |
| `ui/` | Next.js local query Studio and dashboard prototypes. |

## Requirements

The core database requires:

- Go 1.23.5 or newer
- Make
- Git

The local Studio additionally requires:

- Node.js 20.9.0 or newer
- npm

Continuous integration validates the Go project on Ubuntu and macOS. The UI checks run with Node.js 22.

## Quick start

Clone and build the project:

```bash
git clone https://github.com/rushikeshg25/cool-db.git
cd cool-db
make build
```

The compiled binary is written to `bin/cool`.

Start a server with a project-local database file:

```bash
./bin/cool server --db ./data/example.cooldb
```

The server listens on `localhost:3040` by default. It creates the database directory and file when the first mutation is persisted. Stop it with `Ctrl+C`.

In another terminal, create a table and add a row:

```bash
./bin/cool exec "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL, active BOOLEAN)"
./bin/cool exec "INSERT INTO users VALUES (1, 'Ada Lovelace', true)"
./bin/cool exec "SELECT * FROM users"
```

Expected query output:

```text
id   name          active
---  ---           ---
1    Ada Lovelace  true
(1 row(s))
```

Open an interactive session against the same server:

```bash
./bin/cool shell
```

Exit the shell with `quit`, `exit`, `.quit`, `.exit`, or `\q`.

## Command-line reference

Run `./bin/cool --help` or `./bin/cool <command> --help` for the complete generated reference.

| Command | Purpose |
| --- | --- |
| `cool server` | Start the database engine and gRPC server. `cool start` is an alias. |
| `cool shell` | Open an interactive SQL shell. |
| `cool exec [query]` | Execute exactly one SQL statement and print its result. |
| `cool version` | Print the version and build time. |
| `cool completion <shell>` | Generate shell completion instructions. |

### Server options

| Option | Default | Description |
| --- | --- | --- |
| `--host` | `localhost` | Hostname or address used by the server. |
| `--port`, `-p` | `3040` | gRPC port. |
| `--db` | `~/cooldb/default.cooldb` | Database snapshot path. |
| `--http-port` | `0` | Local Studio API port. A value of `0` disables the HTTP bridge. |
| `--wal`, `-w` | Disabled | Reserved for write-ahead logging. Version 0.1 returns an error if it is enabled. |

### Client options

The `shell` and `exec` commands accept the same connection settings:

| Option | Default | Description |
| --- | --- | --- |
| `--host` | `localhost` | CoolDB server hostname or address. |
| `--port`, `-p` | `3040` | CoolDB gRPC port. |
| `--connect-timeout` | `5s` | Maximum time allowed to establish a connection. |
| `--query-timeout` | `30s` | Maximum time allowed for a query. |

Pass SQL directly to `exec`:

```bash
./bin/cool exec "SELECT id, name FROM users WHERE id = 1"
```

Or read one statement from standard input:

```bash
printf 'SELECT * FROM users;\n' | ./bin/cool exec
```

To connect to another server:

```bash
./bin/cool shell --host 127.0.0.1 --port 4040
```

## Supported SQL

CoolDB executes exactly one statement per request. A trailing semicolon is optional. SQL keywords and identifiers are case-insensitive, and identifiers are stored in lowercase.

### Statements

| Operation | Example |
| --- | --- |
| Create a table | `CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT UNIQUE, name VARCHAR(100) NOT NULL)` |
| Drop a table | `DROP TABLE users` |
| Insert a complete row | `INSERT INTO users VALUES (1, 'ada@example.com', 'Ada')` |
| Insert selected columns | `INSERT INTO users (id, name) VALUES (2, 'Grace')` |
| Select all columns | `SELECT * FROM users` |
| Select named columns | `SELECT id, name FROM users WHERE id = 1` |
| Update matching rows | `UPDATE users SET name = 'Ada Lovelace' WHERE id = 1` |
| Delete matching rows | `DELETE FROM users WHERE id = 1` |

`UPDATE` and `DELETE` affect every row when the `WHERE` clause is omitted.

### Data types

| Type | Aliases | Accepted values |
| --- | --- | --- |
| `INTEGER` | `INT` | Signed 64-bit integers. |
| `TEXT` | None | Single-quoted strings. Use two single quotes to represent an apostrophe. |
| `VARCHAR(n)` | None | Text limited to `n` Unicode characters. |
| `BOOLEAN` | `BOOL` | `true` or `false`. |
| `FLOAT` | `DOUBLE` | Integer or decimal numeric literals stored as 64-bit floating-point values. |

All column types accept `NULL` unless the column is declared `NOT NULL` or `PRIMARY KEY`.

### Constraints and query behavior

- A table can have at most one `PRIMARY KEY` column.
- Primary keys are implicitly unique and non-null.
- `UNIQUE` values are checked across all rows in the table.
- `VARCHAR(n)` validates the number of Unicode characters.
- `WHERE` currently supports one equality predicate in the form `column = literal`.
- `SELECT` performs a table scan and returns rows in stored order.
- String literals use SQL-style escaping, such as `'D''Angelo'`.

Joins, indexes, transactions, aggregates, sorting, grouping, schema alterations, and compound predicates are not implemented in version 0.1.

## Persistence model

Each server process owns one database snapshot. If `--db` is omitted, the default path is `~/cooldb/default.cooldb`.

The current storage model has these properties:

- Database state is persisted as a versioned JSON snapshot.
- Each successful mutation is applied to a cloned in-memory state and persisted before it becomes the active state.
- Snapshot replacement uses a temporary file, file synchronization, an atomic rename, and directory synchronization.
- Newly created snapshot files use owner-only `0600` permissions.
- Concurrent access inside one server process is protected with a read-write mutex.
- Reads do not rewrite the database file.

The snapshot format is an implementation detail and should not be edited manually. Stop the server before copying a database file for backup. Online backups and write-ahead logging are not yet available.

## Local Studio

The Studio is a local demonstration interface and design prototype. It can execute both read and write statements against the selected database.

Start the complete demo from the repository root:

```bash
make demo
```

This command:

1. Builds the `cool` binary.
2. Starts the gRPC server on `127.0.0.1:3040`.
3. Starts the opt-in HTTP bridge on `127.0.0.1:3041`.
4. Installs UI dependencies when `ui/node_modules` is absent.
5. Starts Next.js on `http://localhost:3000`.

The default scratch database is `.cooldb-demo/demo.cooldb`, which is ignored by Git.

Open one of the available interface variants:

- [Command center](http://localhost:3000/?variant=A), a dense operational workspace
- [Query notebook](http://localhost:3000/?variant=B), a document-oriented exploration interface
- [Terminal wall](http://localhost:3000/?variant=C), a keyboard-focused console interface

Use the floating switcher or the left and right arrow keys to change variants during development.

### Demo configuration

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `COOLDB_DEMO_GRPC_PORT` | `3040` | gRPC port started by the demo script. |
| `COOLDB_DEMO_HTTP_PORT` | `3041` | Local HTTP bridge port. |
| `COOLDB_DEMO_UI_PORT` | `3000` | Next.js development server port. |
| `COOLDB_DEMO_DB` | `.cooldb-demo/demo.cooldb` | Demo database path. |
| `COOLDB_DEMO_API_URL` | `http://127.0.0.1:3041` | HTTP bridge used when the UI is run separately. |

For example:

```bash
COOLDB_DEMO_UI_PORT=3100 COOLDB_DEMO_DB=/tmp/cooldb-demo.cooldb make demo
```

To run the services separately:

```bash
make build
./bin/cool server \
  --host 127.0.0.1 \
  --port 3040 \
  --http-port 3041 \
  --db ./.cooldb-demo/demo.cooldb
```

Then start the UI in another terminal:

```bash
npm --prefix ui install
npm --prefix ui run dev
```

The HTTP bridge has no authentication and is not a public API. Keep it bound to a loopback address, use a scratch database, and do not expose it to an untrusted network. See [`ui/README.md`](ui/README.md) for UI-specific details.

## Development

### Common commands

| Command | Description |
| --- | --- |
| `make build` | Compile `bin/cool` with version and build-time metadata. |
| `make run` | Build and start the server with default settings. |
| `make shell` | Build and open the interactive shell. A server must already be running. |
| `make demo` | Build and start the server, HTTP bridge, and local Studio. |
| `make test` | Run the Go test suite. |
| `make tidy` | Normalize Go module dependencies. |
| `make clean` | Remove the compiled binary. |

### Full validation

Run the checks used by continuous integration before opening a pull request:

```bash
go test -race ./...
go vet ./...
npm --prefix ui ci
npm --prefix ui audit
npm --prefix ui run lint
npm --prefix ui run build
```

The repository also runs `golangci-lint` for Go changes and a scheduled Trivy filesystem scan.

### Contribution workflow

1. Create a focused branch from `main`.
2. Keep each commit limited to one coherent change.
3. Add or update tests for behavioral changes.
4. Run the full validation commands.
5. Open a pull request with the motivation, implementation details, and verification results.

## Current limitations

- CoolDB is a version 0.1 experimental database and is not production-ready.
- The gRPC server does not provide authentication, authorization, or TLS configuration.
- The local HTTP bridge has no authentication and must remain local.
- Every mutation writes a complete database snapshot.
- The engine does not provide transactions, write-ahead logging, crash recovery through log replay, or online backups.
- Queries use table scans because indexes are not implemented.
- The Studio returns text-formatted results and its three layouts are design prototypes.
- The on-disk format may change before a stable release.

## Roadmap

Planned work includes write-ahead logging, administrative shell commands, indexes, caching, authentication, backup support, schema visualization, and deployment tooling. See [`plan.md`](plan.md) for the current development plan.

## License

This repository does not currently include a software license. A license must be added before the project can be treated as open source or redistributed under explicit terms.
