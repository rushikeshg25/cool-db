# Decisions

These choices are observed in the implementation. Rationale is marked as inferred unless a source comment states it; the older prospective architecture is not used as proof of the current design's intent.

## One engine behind two transport adapters

- **What:** the server opens one Engine and passes it to gRPC and HTTP, with both depending on a narrow `Execute` contract.
- **Evidence:** [server construction](../../server/main.go#L51), [HTTP wiring](../../server/main.go#L79), [core interface](../../internal/core/main.go#L28), [HTTP interface](../../internal/httpapi/handler.go#L16).
- **Why, apparently:** inferred separation of SQL/storage semantics from request protocols; both front ends see the same active state.
- **Tradeoff:** transport policies and error mappings must be maintained twice. HTTP has different request limits and status codes; it does not exercise the gRPC path.
- **Confidence:** wiring confirmed; rationale inferred.

## Whole-state copies make statement failure easy to contain

- **What:** every mutation deep-copies the catalog and rows, edits the candidate, saves it, and only then replaces active state.
- **Evidence:** [Execute](../../internal/database/engine.go#L56), [cloneState](../../internal/database/engine.go#L330), [injected failure test](../../internal/database/engine_test.go#L103).
- **Why, apparently:** inferred simple failure isolation without undo records. A SQL or pre-commit persistence error can discard the candidate.
- **Tradeoff:** mutation memory/CPU/write cost grows with the entire database, including unrelated tables and zero-row mutations. This is statement-level publication, not support for SQL transaction sessions.
- **Confidence:** implementation confirmed; rationale inferred.

## Snapshot replacement is the commit boundary

- **What:** sync the temporary file before rename, then treat directory-sync failure as a warning after rename.
- **Evidence:** [save ordering](../../internal/database/persistence.go#L154) and [explicit commit rationale](../../internal/database/persistence.go#L171).
- **Why:** once rename has succeeded, rejecting the mutation would cause the engine to keep old memory while the file already contains new state.
- **Tradeoff:** memory/file agreement is preserved, but failure to sync the directory can weaken durability across power loss. The tests inject an Engine persistence error; they do not simulate an operating-system crash at every write stage ([failure test](../../internal/database/engine_test.go#L103)).
- **Confidence:** confirmed by code and comment.

## A single reader-writer lock defines concurrency

- **What:** SELECT shares a read lock; every other supported statement takes one exclusive lock through persistence. Row access and uniqueness checking use scans.
- **Evidence:** [locking](../../internal/database/engine.go#L50), [SELECT loop](../../internal/database/engine.go#L183), [unique scan](../../internal/database/engine.go#L312).
- **Why, apparently:** inferred simple consistency model without page locks, transactions, or index maintenance.
- **Tradeoff:** readers can overlap, but all writes serialize and block reads during filesystem work. This lock does not coordinate separate Engine instances or processes sharing a file.
- **Confidence:** behavior confirmed; rationale and scaling implications inferred.

## Tagged JSON values preserve types and integer precision

- **What:** a value has a type tag, explicit NULL marker, and dedicated int64/string/bool/float64 fields; VARCHAR is TEXT with a length constraint.
- **Evidence:** [Value comment and fields](../../internal/database/types.go#L37), [type parsing](../../internal/database/sql.go#L291), [load version check](../../internal/database/persistence.go#L34).
- **Why:** the comment explicitly calls out preserving integer precision through JSON persistence.
- **Tradeoff:** readable snapshots carry more per-cell metadata than a compact page format. Other snapshot versions are rejected rather than migrated; loading checks a defined subset of invariants.
- **Confidence:** precision rationale confirmed; format tradeoffs inferred.

## Structured engine results become text at transport boundaries

- **What:** Engine returns rows/columns/counts; both adapters call `Format`, and clients receive strings.
- **Evidence:** [Result and future protocol comment](../../internal/database/result.go#L9), [wire response](../../internal/core/main.go#L77), [HTTP response](../../internal/httpapi/handler.go#L74).
- **Why:** the formatter comment explicitly keeps open a future structured protocol without changing Engine.
- **Tradeoff:** shell/UI share rendering today, but typed client consumption, sorting, and richer result grids need a protocol change or parsing of display text. Tabs/newlines in stored strings are rendered as text, not a machine-safe interchange format ([formatter](../../internal/database/result.go#L44)).
- **Confidence:** confirmed by code and comment.

## Small handwritten grammar keeps the supported dialect explicit

- **What:** lex rune tokens, parse six statement types, and require EOF after an optional semicolon. No external query planner is involved.
- **Evidence:** [parseSQL](../../internal/database/sql.go#L107), [lexer](../../internal/database/sql.go#L124), [statement dispatch](../../internal/database/sql.go#L216), [single predicate](../../internal/database/sql.go#L435).
- **Why, apparently:** inferred control over a small educational SQL subset.
- **Tradeoff:** each SQL feature requires parser and executor work. Joins, sorting, aggregates, compound predicates, arithmetic, and multi-statement requests do not fit the current grammar.
- **Confidence:** grammar confirmed; rationale inferred.

## Test seams sit at side-effect boundaries

- **What:** commands inject server/client/input factories; the shell consumes Executor/LineReader interfaces; adapters consume a narrow engine interface.
- **Evidence:** [command dependencies](../../cmd/root.go#L19), [shell contracts](../../internal/shell/shell.go#L18), [command fakes](../../cmd/query_test.go#L14), [bufconn client tests](../../internal/client/client_test.go#L27).
- **Why, apparently:** inferred ability to test CLI behavior without real terminals and client behavior without external services, while retaining a real restart workflow test.
- **Tradeoff:** constructors have dependency plumbing; fakes do not prove whole-system behavior, so the [integration test](../../cmd/integration_test.go#L16) remains useful.
- **Confidence:** test usage confirmed; intent inferred.

## Studio variants share behavior but own separate state

- **What:** A/B/C use one hook implementation, each with its own instance; the dashboard swaps component types. The source calls the layouts a throwaway design prototype.
- **Evidence:** [prototype comment and selector](../../ui/app/_demo/demo-dashboard.tsx#L24), [component render](../../ui/app/_demo/demo-dashboard.tsx#L66), [hook state](../../ui/app/_demo/use-demo-query.ts#L30), [variant hook call](../../ui/app/_demo/variant-notebook.tsx#L7).
- **Why, apparently:** explicit prototype status; inferred goal of comparing presentations over the same query behavior.
- **Tradeoff:** switching variants resets query/result state. The schema panels and system labels are authored display content, not live database introspection.
- **Confidence:** prototype intent confirmed; state-reset consequence inferred from component ownership.

## The UI Go module is an intentional tooling fence

- **What:** `ui/go.mod` defines a nested module with no Go application source.
- **Evidence:** [full explanatory comment](../../ui/go.mod#L1).
- **Why:** stop root build/vet/test traversal at `ui/`, excluding Go files vendored under npm dependencies.
- **Tradeoff:** it can look like an accidental extra service/module; deleting it changes Go tooling traversal.
- **Confidence:** explicitly documented.

## Gotchas

- **Old specs overstate the implementation.** The [design overview](../../.kiro/specs/sql-database-engine/design.md#L3) describes B+ trees, ACID transactions, WAL and a database/sql driver; current [execution](../../internal/database/engine.go#L44) is an in-memory catalog with JSON snapshots. [plan.md](../../plan.md#L25) also leaves shell exits, gRPC statuses, locks, and CI in future sections despite current code.
- **The README's NULL statement has an exception.** [README.md:218](../../README.md#L218) says `column = NULL` never matches; nullable columns behave that way, but NOT NULL/primary-key columns currently raise a constraint error because the matcher uses assignment coercion ([matcher](../../internal/database/engine.go#L305), [NULL coercion](../../internal/database/types.go#L92)). Use `IS NULL`/`IS NOT NULL` for null checks.
- **Timeout does not mean rollback.** Client/proxy timeouts stop waiting, while an engine call already entered can finish and persist ([client deadline](../../internal/client/client.go#L79), [core context check](../../internal/core/main.go#L69), [HTTP execution](../../internal/httpapi/handler.go#L69)).
- **Local demo is intended scope, not enforced access control.** Configured host reaches both listeners; Next's launch does not set a loopback hostname. There is no auth middleware ([HTTP binding](../../server/main.go#L74), [gRPC setup](../../internal/core/main.go#L47), [demo launch](../../scripts/demo.sh#L44)).
- **Studio transport label is stale.** The [Command center card](../../ui/app/_demo/variant-command-center.tsx#L158) says `HTTP → gRPC`; the [Go handler](../../internal/httpapi/handler.go#L69) calls Engine directly. The displayed database/schema/version cards are static too ([explorer](../../ui/app/_demo/variant-command-center.tsx#L49), [terminal schema](../../ui/app/_demo/variant-terminal.tsx#L98)).
- **A fresh demo has no users table.** The hook starts with a SELECT; sample buttons set editor text rather than seed the database. Execute Create table, then Insert Ada, then Browse users ([hook](../../ui/app/_demo/use-demo-query.ts#L5), [button](../../ui/app/_demo/variant-command-center.tsx#L71)).
- **Keyboard behavior differs by layout.** A/C implement Cmd/Ctrl+Enter; B only has an evaluate button. Terminal F1–F4 labels have no function-key handlers. Only the floating switcher is hidden in production; dashboard arrow navigation remains ([A handler](../../ui/app/_demo/variant-command-center.tsx#L115), [B editor](../../ui/app/_demo/variant-notebook.tsx#L82), [C labels/handler](../../ui/app/_demo/variant-terminal.tsx#L34), [dashboard listener](../../ui/app/_demo/demo-dashboard.tsx#L48), [production guard](../../ui/app/_demo/demo-dashboard.tsx#L85)).
- **Health and timing are narrow signals.** Health is a fixed response, and the displayed query duration includes browser/proxy/network overhead rather than server execution alone ([health](../../internal/httpapi/handler.go#L46), [timer](../../ui/app/_demo/use-demo-query.ts#L72)).
- **Shell is line-oriented; exec reads all stdin as one statement.** Piping a multi-statement SQL file does not provide script execution. Unknown `.tables`/`.open`-style commands are rejected locally ([shell parser/loop](../../internal/shell/shell.go#L40), [stdin](../../cmd/query.go#L98), [SQL EOF check](../../internal/database/sql.go#L117)).
- **Dangerous omissions are valid SQL here.** Omitting WHERE from UPDATE/DELETE affects all rows; there is no confirmation inside Engine ([matcher](../../internal/database/engine.go#L289)). UNIQUE permits multiple NULLs, and VARCHAR length counts runes rather than bytes ([row validation](../../internal/database/engine.go#L318), [length check](../../internal/database/types.go#L114)).
- **Flags can advertise future functionality.** `--wal` always errors when enabled; nonpositive client timeout values restore defaults rather than disable deadlines ([WAL](../../cmd/server.go#L27), [client defaults](../../internal/client/client.go#L100)).

## Conventions

- Add categorized database errors and map them at transport boundaries; do not make clients parse engine error messages ([error categories](../../internal/database/errors.go#L5), [gRPC mapping](../../internal/core/main.go#L80), [HTTP mapping](../../internal/httpapi/handler.go#L88)).
- Keep SQL semantics in the database package, and test there with temporary files and table-driven cases ([constraint tests](../../internal/database/engine_test.go#L65), [snapshot tests](../../internal/database/persistence_test.go#L20)). Exercise externally visible behavior separately through adapter and restart tests ([HTTP tests](../../internal/httpapi/handler_test.go#L35), [CLI integration](../../cmd/integration_test.go#L16)).
- Use the shared Studio hook for query behavior and keep layout changes in variant components ([shared hook](../../ui/app/_demo/use-demo-query.ts#L30), [variant imports](../../ui/app/_demo/variant-command-center.tsx#L3)).

Return to the [guide index](README.md) or [file map](03-structure.md).
