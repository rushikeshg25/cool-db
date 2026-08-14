# Cooldb

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23.5-blue.svg)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-v15-black.svg)](https://nextjs.org)

**Cooldb** is a full-stack database ecosystem featuring a high-performance gRPC server, a modern web dashboard, and a robust CLI client. It is designed for developers who need a lightweight, extensible, and "cool" way to manage their data.

## 🚀 Key Features

- **gRPC Infrastructure**: Built on a high-performance communication layer via `cool-wire`.
- **Integrated Dashboard**: Real-time data visualization built with Next.js.
- **Interactive CLI**: Rich terminal experience with readline support.
- **SQL-like Syntax**: Intuitive query interface for data management.
- **Extensible Architecture**: Clean internal structure for adding custom storage engines or protocols.

## 🏗 Architecture

Cooldb is split into three core components:

| Component | Description | Technology |
| :--- | :--- | :--- |
| **Server** | Core engine handling query execution and storage. | Go, gRPC |
| **CLI Client** | Terminal-based interactive management tool. | Go, Cobra, Readline |
| **UI Dashboard** | Web interface for visual database administration. | Next.js, Tailwind CSS |

For more details on the technical roadmap and future plans, see [plan.md](file:///Users/rushikeshghotekar/Projects/cool-db/plan.md).

## 🚦 Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (1.23.5 or higher)
- [Node.js](https://nodejs.org/) (for the UI)
- [Makefile](https://www.gnu.org/software/make/)

### Running the Server

```bash
make run
```

### Launching the Dashboard

```bash
cd ui
npm install
npm run dev
```

### Using the CLI

Clone and build the [cool-cli](file:///Users/rushikeshghotekar/Projects/cool-cli/README.md) repository:

```bash
# In the cool-cli directory
make build
./bin/cool-cli
```

## 🛠 Development

### Directory Structure

- `cmd/`: Command-line entry points for the server.
- `internal/`: Core logic including storage, core engine, and error handling.
- `server/`: Server initialization and gRPC setup.
- `ui/`: Next.js application source code.

---
Built with ❤️ by [rushikeshg25](https://github.com/rushikeshg25).
