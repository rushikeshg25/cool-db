# Cooldb

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23.5-blue.svg)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-v15-black.svg)](https://nextjs.org)

**Cooldb** is a small SQL database with a persistent Go engine, a gRPC server, and a unified command-line client.

## 🚀 Key Features

- **gRPC Infrastructure**: Built on a high-performance communication layer via `cool-wire`.
- **Persistent Storage**: Typed tables are saved atomically and survive server restarts.
- **Unified CLI**: Run the server, open an interactive shell, or execute scripted queries with one binary.
- **SQL Syntax**: Supports `CREATE TABLE`, `DROP TABLE`, `INSERT`, `SELECT`, `UPDATE`, and `DELETE`.
- **Extensible Architecture**: Clean internal structure for adding custom storage engines or protocols.

## 🏗 Architecture

Cooldb ships as one `cool` binary with three commands:

| Component | Description | Technology |
| :--- | :--- | :--- |
| **`cool server`** | Runs the database engine and gRPC transport. | Go, gRPC |
| **`cool shell`** | Opens an interactive SQL terminal. | Go, Cobra, Readline |
| **`cool exec`** | Executes one query for scripts and CI. | Go, Cobra |

The experimental dashboard remains in `ui/`. For future database work, see [plan.md](plan.md).

## 🚦 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (1.23.5 or higher)
- [Node.js](https://nodejs.org/) (for the UI)
- [Makefile](https://www.gnu.org/software/make/)

### Build and run the server

```bash
make build
./bin/cool server --db ./example.cooldb
```

`cool start` remains an alias for `cool server` during the CLI migration.

### Open the interactive shell

```bash
./bin/cool shell --host localhost --port 3040
```

Exit with `quit`, `exit`, `.quit`, `.exit`, or `\q`.

### Execute a scripted query

Pass the query as an argument:

```bash
./bin/cool exec "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)"
./bin/cool exec "INSERT INTO users VALUES (1, 'Ada')"
./bin/cool exec "SELECT * FROM users"
```

Or pipe a query through standard input:

```bash
printf 'SELECT * FROM users;\n' | ./bin/cool exec
```

### Launch the experimental dashboard

```bash
cd ui
npm install
npm run dev
```

## 🛠 Development

### Directory Structure

- `cmd/`: Unified `server`, `shell`, and `exec` commands.
- `internal/client/`: Transport-independent CoolDB client.
- `internal/shell/`: Interactive shell and terminal adapter.
- `internal/database/`: SQL parsing, validation, execution, and persistence.
- `internal/core/`: gRPC server adapter.
- `server/`: Server initialization and gRPC setup.
- `ui/`: Experimental Next.js application.

The former `cool-cli` repository is no longer required for new development. It can remain available temporarily for users of the standalone binary.

---
Built with ❤️ by [rushikeshg25](https://github.com/rushikeshg25).
