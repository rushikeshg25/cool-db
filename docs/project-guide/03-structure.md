# Structure

## What lives where

The root builds one Go executable. `cmd/` owns user-facing commands, `server/` owns process lifecycle, and `internal/` separates transport, terminal, and database concerns. `ui/` is a separate Next.js application whose server routes call Go's HTTP bridge; its nested Go module exists only to stop root Go tooling from traversing npm dependencies ([root entry](../../main.go#L5), [command wiring](../../cmd/root.go#L41), [UI module comment](../../ui/go.mod#L1)).

```text
main.go                    Go entry point
cmd/                       Cobra commands and workflow tests
server/                    engine/transport lifecycle
internal/
  client/                  outgoing gRPC client
  core/                    incoming gRPC adapter
  database/                parser, execution, types, snapshots
  httpapi/                 optional HTTP demo bridge
  shell/                   interactive loop and terminal adapter
ui/
  app/                     Next page, layout, styling
    _demo/                 variants and shared query hook
    api/health/            health proxy route
    api/query/             SQL proxy route
  lib/                     server-side upstream HTTP helper
scripts/                   local demo orchestration
.github/workflows/         build, tests, lint, scanning
.kiro/specs/               prospective database design
```

Each source link anchors the definition or behavior described. Private functions are included under key exports when they are the useful navigation point.

## Repository root

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [main.go](../../main.go#L5) | Enter the command tree. | `main` | Go runtime |
| [go.mod](../../go.mod#L1) | Root module, Go version, Go dependencies. | Module `github.com/rushikeshg25/coolDb` | Go tooling |
| [makefile](../../makefile#L2) | Binary metadata/build and local task shortcuts. | `build`, `run`, `shell`, `demo`, `test`, `tidy`, `clean` | Developer, demo script |
| [README.md](../../README.md#L1) | User instructions and supported scope. | Documentation | Maintainers/users |
| [plan.md](../../plan.md#L7) | Phased roadmap; partly stale. | Planning sections | Planning work; no runtime caller |

## cmd

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [root.go](../../cmd/root.go#L14) | Isolated command tree, injected runtime dependencies, version, error exit. | `Version`, `BuildTime`, `NewRootCommand`, `Execute`; private `dependencies` | Root main, tests |
| [server.go](../../cmd/server.go#L13) | Server flags, WAL rejection, signal context. | `newServerCommand` | Root constructor |
| [query.go](../../cmd/query.go#L14) | Shared client flags, shell/exec construction, SQL argument/stdin extraction. | `connectionFlags`, `newShellCommand`, `newExecCommand`, `queryFromInput` | Root constructor |
| [root_test.go](../../cmd/root_test.go#L25) | Help/version, server config forwarding, WAL rejection. | `TestRootShowsHelp`, `TestStartPassesConfigurationToServer` | Go test |
| [query_test.go](../../cmd/query_test.go#L14) | Fake client/input tests for shell and exec input. | `fakeManagedClient`, `fakeInteractiveInput`, `TestExecAcceptsStdin` | Go test |
| [integration_test.go](../../cmd/integration_test.go#L16) | Real gRPC workflow and snapshot persistence across restart. | `TestUnifiedExecWorkflowPersistsAcrossRestart` | Go test |

## server

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [main.go](../../server/main.go#L20) | Resolve database path, open engine, run one/two transports, join shutdown. | `Config`, `Run`; private `runWithHTTP`, `normalizeHTTPError` | Command production dependencies, integration tests |
| [main_test.go](../../server/main_test.go#L14) | HTTP startup health and context-driven shutdown. | `TestRunStartsAndStopsDemoHTTPServer` | Go test |

## internal/client

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [client.go](../../internal/client/client.go#L25) | gRPC connection readiness, deadlines, response/error conversion. | `Config`, `Client`, `Connect`, `Execute`, `Close`, `Result`, `QueryError` | Commands; shell consumes client result type |
| [client_test.go](../../internal/client/client_test.go#L27) | In-memory gRPC service tests for results, errors, deadlines. | `connectTestClient`, `TestExecuteHonorsTimeout` | Go test |

## internal/core

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [main.go](../../internal/core/main.go#L17) | Bind gRPC, register wire service, adapt requests and database errors. | `CoreServer`, `CoreServerGRPC`, `NewCoreServer`, `BindAndListen`, `SendQuery` | Server lifecycle; gRPC dispatch |
| [main_test.go](../../internal/core/main_test.go#L15) | Adapter behavior with real engine and status mapping. | `TestSendQueryExecutesAgainstDatabase`, `TestSendQueryMapsDatabaseErrorsToGRPC` | Go test |

## internal/database

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [engine.go](../../internal/database/engine.go#L9) | State ownership, locking, mutation dispatch, scans, constraints, cloning. | `Engine`, `Open`, `Execute`; private `databaseState`, `table`, `executeSelect`, `validateRow` | Server, transport adapters, tests |
| [sql.go](../../internal/database/sql.go#L107) | Handwritten lexer/parser and internal statement types. | Private `parseSQL`, `lex`, `parser`, statement structs | Engine execution |
| [types.go](../../internal/database/types.go#L8) | Typed cells/schema, coercion, equality, text rendering. | `DataType`, type constants, `Column`, `Value`, `Value.String` | Parser, engine, persistence, result formatter |
| [persistence.go](../../internal/database/persistence.go#L14) | Load/validate snapshots, save through temporary-file replacement. | Private `loadState`, `validateState`, `validateTable`, `saveState`, `syncDirectory` | Engine Open and persistence closure |
| [result.go](../../internal/database/result.go#L9) | Structured execution outcome and human output. | `Result`, `Format` | Engine creates; transports format |
| [errors.go](../../internal/database/errors.go#L5) | Stable categories, messages, wrapped causes. | `ErrorCode`, code constants, `Error`, `Unwrap` | Database helpers; transport mapping |
| [engine_test.go](../../internal/database/engine_test.go#L11) | CRUD/reopen, constraints, injected persistence failure, formatting, NULL and exponent semantics. | `TestEngineCRUDAndPersistence`, `TestFailedPersistenceDoesNotMutateEngine`, further `Test*` cases | Go test |
| [persistence_test.go](../../internal/database/persistence_test.go#L20) | Invalid snapshots, short-row regression, valid load, untyped NULL normalization. | `TestOpenRejectsCorruptSnapshots`, `TestOpenNormalizesUntypedNulls` | Go test |

## internal/httpapi

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [handler.go](../../internal/httpapi/handler.go#L14) | Health/query routes, JSON size/shape checks, error statuses, response headers. | `Handler`, `NewHandler`; private request/response types and `queryExecutor` | Server HTTP bootstrap |
| [handler_test.go](../../internal/httpapi/handler_test.go#L14) | Real temporary engine with httptest requests for health, CRUD, error payloads, invalid JSON fields. | `newTestHandler`, `TestHealth`, `TestQueryExecutesSQL`, `TestQueryRejectsInvalidBody` | Go test |

## internal/shell

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [shell.go](../../internal/shell/shell.go#L16) | Input classification, execution loop, signal context, rendering. | `Executor`, `LineReader`, `InputKind`, `Input`, `Parse`, `Runner`, `ErrInterrupted`, `RenderResult`, `RenderError` | Shell/exec commands |
| [readline.go](../../internal/shell/readline.go#L12) | Terminal library wrapper with history and interrupt translation. | `Readline`, `NewReadline`, `Readline.Readline`, `Close` | Command input factory |
| [shell_test.go](../../internal/shell/shell_test.go#L14) | Classification, visible output/error continuation, interrupts, newline behavior. | `TestParse`, `TestRunnerExecutesQueriesAndRendersVisibleBehavior` | Go test |

## scripts

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [demo.sh](../../scripts/demo.sh#L5) | Build, conditional npm install, Go child lifecycle, Next foreground process. | Environment configuration, `cleanup`, `handle_signal` | `make demo` |

## ui

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [package.json](../../ui/package.json#L5) | Node floor, npm commands, UI dependency ranges/pins. | `dev`, `build`, `start`, `lint` | npm, script, CI |
| [go.mod](../../ui/go.mod#L1) | Fence off npm dependencies from root Go tooling. | Empty nested Go module | Go tooling |
| [next.config.ts](../../ui/next.config.ts#L3) | Disable framework-generated agent instruction files. | Default `nextConfig` | Next |
| [tsconfig.json](../../ui/tsconfig.json#L2) | Strict TS, bundler resolution, Next plugin, UI-root import alias. | Compiler configuration | TypeScript/Next |
| [eslint.config.mjs](../../ui/eslint.config.mjs#L1) | Next web-vitals/TS presets and generated-file ignores. | Default ESLint configuration | `npm run lint` |
| [postcss.config.mjs](../../ui/postcss.config.mjs#L1) | Tailwind CSS transform. | Default PostCSS configuration | Next CSS pipeline |
| [README.md](../../ui/README.md#L1) | Studio usage, prototype status, separate-process configuration. | Documentation | UI developers |

## ui/app

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [layout.tsx](../../ui/app/layout.tsx#L5) | Geist fonts, metadata, document wrapper, global CSS. | `metadata`, default `RootLayout` | Next App Router |
| [page.tsx](../../ui/app/page.tsx#L5) | Dashboard root and Suspense fallback. | Default `Home` | Next root route |
| [globals.css](../../ui/app/globals.css#L1) | Tailwind import, color/font variables, baseline controls. | CSS theme and global rules | Root layout |

## ui/app/_demo

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [demo-dashboard.tsx](../../ui/app/_demo/demo-dashboard.tsx#L18) | URL variant selection, arrow navigation, development switcher. | `DemoDashboard`; private `PrototypeSwitcher` | Home |
| [use-demo-query.ts](../../ui/app/_demo/use-demo-query.ts#L5) | Example SQL, query/result/status state, health fetch, query timing. | `demoQueries`, `useDemoQuery` | All three variants |
| [variant-command-center.tsx](../../ui/app/_demo/variant-command-center.tsx#L5) | Dark explorer layout, reconnect, keyboard execution, static session cards. | `variantName`, `VariantCommandCenter` | Dashboard, Home fallback |
| [variant-notebook.tsx](../../ui/app/_demo/variant-notebook.tsx#L5) | Notebook editor, evaluation button, output, static notes. | `variantName`, `VariantNotebook` | Dashboard |
| [variant-terminal.tsx](../../ui/app/_demo/variant-terminal.tsx#L5) | Terminal editor, keyboard execution, output, static schema map. | `variantName`, `VariantTerminal`; private `Section`, `TerminalRow` | Dashboard |

## ui/app/api/query

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [route.ts](../../ui/app/api/query/route.ts#L3) | Validate browser query JSON, proxy Go request, preserve response status. | `POST` | Next route dispatch from query hook |

## ui/app/api/health

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [route.ts](../../ui/app/api/health/route.ts#L3) | Proxy health JSON/status or return unavailable response. | `GET` | Next route dispatch from query hook |

## ui/lib

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [cooldb.ts](../../ui/lib/cooldb.ts#L1) | Upstream URL, no-store fetch with timeout, 503 response. | `DemoApiError`, `requestDemoApi`, `unavailableResponse` | Both Next API routes |

## .github/workflows

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [test.yml](../../.github/workflows/test.yml#L3) | Go build matrix/race/vet; UI install/audit/lint/build. | `build`, `go`, `ui` jobs | Push/PR targeting main |
| [lint.yml](../../.github/workflows/lint.yml#L3) | Go-file-filtered golangci-lint run. | `golangci` job | Push/PR targeting main with Go changes |
| [scan.yml](../../.github/workflows/scan.yml#L3) | Scheduled/manual filesystem vulnerability scan. | `trivy_scan` job | GitHub schedule or dispatch |

## .kiro/specs/sql-database-engine

These are prospective design artifacts, not a description of the current snapshot engine.

| File | Responsibility | Key exports or definitions | Called by |
| --- | --- | --- | --- |
| [design.md](../../.kiro/specs/sql-database-engine/design.md#L3) | Proposed B+ tree, transaction, WAL, buffer-pool architecture. | Design interfaces | No runtime caller |
| [requirements.md](../../.kiro/specs/sql-database-engine/requirements.md#L3) | Desired database capability/acceptance criteria. | Numbered requirements | Planning |
| [tasks.md](../../.kiro/specs/sql-database-engine/tasks.md#L1) | Implementation checklist for that proposed design. | Task list | Planning |

## Excluded

Lockfiles (`go.sum`, `ui/package-lock.json`), Git ignore rules, generated/build output, dependency trees, the favicon, and template SVGs in `ui/public/` are omitted from detailed tables. They do not add a runtime component; the application layout and variant files above contain the authored UI. The six files in this guide document the source rather than participate in its runtime.
