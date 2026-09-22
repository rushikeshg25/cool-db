# CoolDB project guide

> Generated: 2026-09-22 from commit `81a19f6`. Scope: all 65 tracked files, with Go, TypeScript/TSX, CSS, shell, and build/CI configuration. Source anchors refer to this revision.

## What this is

CoolDB implements a small SQL database with typed tables and persistent JSON snapshots ([engine](../../internal/database/engine.go#L22)). It is intended for learning, experimentation, and local demonstrations ([project scope](../../README.md#L9)). One Go binary provides a server and gRPC clients, while a separate Next.js Studio reaches the same engine through an optional HTTP bridge ([commands](../../cmd/root.go#L69), [server](../../server/main.go#L51)).

## The five-file tour

Read these in order to follow a query from the user to durable storage.

| # | File | Why this one | Then look at |
| --- | --- | --- | --- |
| 1 | [cmd/query.go](../../cmd/query.go#L69) | Turns SQL arguments or stdin into a client call. | [CLI query flow](02-flow.md#cli-query-lifecycle) |
| 2 | [internal/client/client.go](../../internal/client/client.go#L74) | Carries SQL over gRPC and returns text or a typed client error. | [Contracts](01-architecture.md#boundaries-and-contracts) |
| 3 | [internal/core/main.go](../../internal/core/main.go#L65) | Adapts the wire request to the engine. | [Core folder](03-structure.md#internalcore) |
| 4 | [internal/database/engine.go](../../internal/database/engine.go#L44) | Parses, locks, executes, and publishes state. | [Read and mutation flow](02-flow.md#engine-read-and-mutation-flow) |
| 5 | [internal/database/persistence.go](../../internal/database/persistence.go#L130) | Establishes the snapshot commit boundary. | [Storage decisions](05-decisions.md#snapshot-replacement-is-the-commit-boundary) |

## Run it

From the repository root, with Go 1.23.5 or newer and Make available ([Go directive](../../go.mod#L3), [build recipe](../../makefile#L5)):

```bash
make build
./bin/cool server --db ./data/example.cooldb
```

Keep that server running; in another terminal:

```bash
./bin/cool exec "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL, active BOOLEAN)"
./bin/cool exec "INSERT INTO users VALUES (1, 'Ada Lovelace', true)"
./bin/cool exec "SELECT * FROM users"
./bin/cool shell
go test ./...
```

These statements follow the [implemented grammar](../../internal/database/sql.go#L216); use a fresh database for the first two. The shell exits on `quit`, `exit`, `.quit`, `.exit`, `\q`, or EOF ([shell parser](../../internal/shell/shell.go#L40)).

For the Studio, install Node.js >=20.9.0 and npm ([manifest](../../ui/package.json#L5)), then run:

```bash
make demo
```

Open [the Studio](http://localhost:3000/?variant=A). The script builds the binary, installs UI dependencies if absent, starts Go on loopback with gRPC 3040 and HTTP 3041, and starts Next.js on port 3000 using `.cooldb-demo/demo.cooldb` ([script](../../scripts/demo.sh#L5)). The sample query initially selects `users`; run the Create table and Insert Ada examples first on a fresh database ([examples and initial state](../../ui/app/_demo/use-demo-query.ts#L5)).

No environment variable is required. Optional script settings are `COOLDB_DEMO_GRPC_PORT`, `COOLDB_DEMO_HTTP_PORT`, `COOLDB_DEMO_UI_PORT`, and `COOLDB_DEMO_DB`; a separately run Studio uses `COOLDB_DEMO_API_URL` ([script](../../scripts/demo.sh#L6), [proxy configuration](../../ui/lib/cooldb.ts#L1)). See [deployment and configuration](01-architecture.md#deployment) for binding and lifecycle details.

## Reading order for this guide

1. [Architecture](01-architecture.md) — components, contracts, data model, and failure boundaries.
2. [Flow](02-flow.md) — startup, CLI, engine, Studio, shutdown, and build paths.
3. [Structure](03-structure.md) — source file responsibilities, exports, callers, and tests.
4. [Tech stack](04-tech-stack.md) — declared versions, dependencies, configuration, and CI.
5. [Decisions](05-decisions.md) — evidence-backed choices, tradeoffs, conventions, and gotchas.

## Open questions

- Which parts of the older B+ tree/WAL/transaction design remain intended? The [design overview](../../.kiro/specs/sql-database-engine/design.md#L3) describes those components, while the implemented [engine](../../internal/database/engine.go#L44) uses whole-state snapshots and [the CLI rejects WAL](../../cmd/server.go#L27). The [roadmap](../../plan.md#L25) also lists already implemented shell exits, gRPC error mapping, and reader/writer locks as future work.
- Should snapshot loading revalidate uniqueness, text length, and addressable table names? [Current validation](../../internal/database/persistence.go#L49) checks structural/type/nullability invariants but does not cover all SQL constraints. Treat the snapshot as an implementation format, not an import API.
- What should be guaranteed when a request times out after execution starts? [Core](../../internal/core/main.go#L69) checks cancellation before calling an engine API that accepts no context. There is no rollback-on-client-timeout contract.
- What is the intended compatibility policy for the external wire schema and future snapshot versions? The [wire dependency](../../go.mod#L7) pins a separate repository, and [loading](../../internal/database/persistence.go#L34) rejects other format versions without migration.

Known documentation/UI discrepancies are recorded in [Gotchas](05-decisions.md#gotchas). This guide describes source behavior; it does not claim a fresh application test run or power-loss validation.
