# Tech stack

## Languages and runtimes

Versions below are declared pins or ranges, not claims about a local installation.

| Language or runtime | Version or constraint | Declared at |
| --- | --- | --- |
| Go | `1.23.5` directive | [go.mod:3](../../go.mod#L3); also [UI tooling boundary](../../ui/go.mod#L6) |
| Node.js | `>=20.9.0`; CI selects major 22 | [ui/package.json:5](../../ui/package.json#L5), [test.yml:44](../../.github/workflows/test.yml#L44) |
| TypeScript / TSX | `^5`; strict, ES2017 target, bundler module resolution | [ui/package.json:27](../../ui/package.json#L27), [ui/tsconfig.json:2](../../ui/tsconfig.json#L2) |
| POSIX shell | `/bin/sh`; no version pin | [scripts/demo.sh:1](../../scripts/demo.sh#L1) |
| CSS | Tailwind v4 pipeline; no separate language version | [ui/app/globals.css:1](../../ui/app/globals.css#L1), [ui/package.json:26](../../ui/package.json#L26) |

## Frameworks and major libraries

| Library | Declared version | Used for | Use site |
| --- | --- | --- | --- |
| Cobra | `v1.9.1` ([pin](../../go.mod#L8)) | Command tree, flags, help. | [cmd/root.go:53](../../cmd/root.go#L53) |
| chzyer/readline | `v1.5.1` ([pin](../../go.mod#L6)) | Terminal line input and history. | [internal/shell/readline.go:21](../../internal/shell/readline.go#L21) |
| cool-wire | `v0.0.0-20250328192440-ca2bfbadcdc3` ([pin](../../go.mod#L7)) | External generated query/response and gRPC service contract. | [internal/core/main.go:23](../../internal/core/main.go#L23), [client.go:81](../../internal/client/client.go#L81) |
| gRPC Go | `v1.71.0` ([pin](../../go.mod#L9)) | TCP RPC server/client, connectivity, statuses, deadlines. | [core/main.go:40](../../internal/core/main.go#L40), [client/client.go:57](../../internal/client/client.go#L57) |
| Protobuf Go | `v1.36.6`, indirect ([pin](../../go.mod#L20)) | Wire serialization dependency used through generated wire types. | [wire import](../../internal/core/main.go#L10) |
| Go standard library | Root Go directive | HTTP, JSON, filesystem sync/rename, mutexes, tabular formatting. | [handler imports](../../internal/httpapi/handler.go#L3), [persistence imports](../../internal/database/persistence.go#L3), [result formatter](../../internal/database/result.go#L20) |
| Next.js | `16.3.3` ([pin](../../ui/package.json#L15)) | App Router pages/API routes, server proxy, build/dev server, fonts. | [page](../../ui/app/page.tsx#L5), [query route](../../ui/app/api/query/route.ts#L3), [layout fonts](../../ui/app/layout.tsx#L2) |
| React / React DOM | `^19.0.0` ([ranges](../../ui/package.json#L16)) | Components, Suspense, state/effects. | [page](../../ui/app/page.tsx#L1), [query hook](../../ui/app/_demo/use-demo-query.ts#L3) |
| Tailwind CSS / PostCSS plugin | `^4` ([ranges](../../ui/package.json#L20)) | Utility classes and theme compilation. | [postcss.config.mjs:1](../../ui/postcss.config.mjs#L1), [globals.css:1](../../ui/app/globals.css#L1) |
| ESLint / Next ESLint config | `9.39.5` / `^16.3.3` ([declarations](../../ui/package.json#L24)) | UI static checks using web-vitals and TypeScript presets. | [eslint.config.mjs:1](../../ui/eslint.config.mjs#L1) |

Go transitive dependencies appear in the [second require block](../../go.mod#L12); npm's exact resolution is maintained in the lockfile. The SQL parser and storage engine use Go code in this repository, rather than an external SQL database library ([parser](../../internal/database/sql.go#L107), [persistence](../../internal/database/persistence.go#L130)).

## Data and infrastructure

| Facility | Role | Configured or implemented at |
| --- | --- | --- |
| Local JSON snapshot | Entire catalog and rows, format version 1 | [engine.go:9](../../internal/database/engine.go#L9), [persistence.go:130](../../internal/database/persistence.go#L130) |
| gRPC listener | Primary shell/exec transport; localhost:3040 default | [CLI flags](../../cmd/server.go#L41), [listener](../../internal/core/main.go#L40) |
| Optional Go HTTP listener | Studio health/query API; shares configured Go host | [server.main.go:59](../../server/main.go#L59), [HTTP bind](../../server/main.go#L73) |
| Next server-side HTTP proxy | Browser same-origin API to configured Go upstream | [ui/lib/cooldb.ts:1](../../ui/lib/cooldb.ts#L1) |
| Local demo processes | Go + Next, no container orchestration | [scripts/demo.sh:18](../../scripts/demo.sh#L18) |

There is no external database service, queue, or cache in the implemented server wiring ([server.Run](../../server/main.go#L28)). Docker is a roadmap item ([plan.md:70](../../plan.md#L70)); no deployment manifest is present among the tracked files at this revision.

## Tooling

| Tool/workflow | Role | Configuration and version |
| --- | --- | --- |
| Make + Go compiler | Builds `bin/cool`; injects version `0.1.0` and UTC build time | [makefile:2](../../makefile#L2); Make itself is not pinned |
| Go test | Package tests; restart integration path, adapter tests and persistence regressions | [makefile:17](../../makefile#L17), [integration test](../../cmd/integration_test.go#L16), [persistence tests](../../internal/database/persistence_test.go#L20) |
| Build CI | Ubuntu/macOS matrix, Go from module directive | [test.yml:13](../../.github/workflows/test.yml#L13), checkout `v4`, setup-go `v5` |
| Go verification CI | `go test -race ./...`, `go vet ./...` on Ubuntu | [test.yml:26](../../.github/workflows/test.yml#L26) |
| UI verification CI | Node 22; `npm ci`, `npm audit`, lint and production build | [test.yml:37](../../.github/workflows/test.yml#L37), setup-node `v4` |
| golangci-lint | Go-file-filtered main push/PR lint workflow | [lint.yml:3](../../.github/workflows/lint.yml#L3); action `v6`, linter `v1.64`, moving `stable` Go at [line 20](../../.github/workflows/lint.yml#L20) |
| Trivy | Monday 00:00 UTC and manual filesystem vulnerability scan; fail on fixed HIGH/CRITICAL findings | [scan.yml:3](../../.github/workflows/scan.yml#L3), action `0.28.0` and options at [line 18](../../.github/workflows/scan.yml#L18) |
| Next/Turbopack | Dev server; separate Next production build/start commands | [ui/package.json:8](../../ui/package.json#L8) |
| Next font integration | Geist and Geist Mono via Google font loader | [ui/app/layout.tsx:2](../../ui/app/layout.tsx#L2) |

For the CI verification path, run from the repository root:

```bash
go build ./...
go test -race ./...
go vet ./...
npm --prefix ui ci
npm --prefix ui audit
npm --prefix ui run lint
npm --prefix ui run build
```

Those commands mirror [test.yml](../../.github/workflows/test.yml#L24). Lint and scheduled scanning are separate workflows. There is no UI behavioral/browser-test script in the [npm scripts](../../ui/package.json#L8), and the workflow does not add one.

## Notes

- The UI's nested Go module is a tooling exclusion boundary, not a second Go service ([explicit comment](../../ui/go.mod#L1)).
- `agentRules: false` prevents generated framework agent instructions; TypeScript's `@/*` alias resolves from the UI root ([Next configuration](../../ui/next.config.ts#L3), [TS alias](../../ui/tsconfig.json#L25)).
- Go lint uses the moving stable toolchain while build/test use the module directive. Passing one does not imply identical toolchain behavior in the other ([lint](../../.github/workflows/lint.yml#L20), [test](../../.github/workflows/test.yml#L20)).
- The Google font loader introduces a build-environment dependency on font acquisition; there are no checked-in local font inputs in [layout.tsx](../../ui/app/layout.tsx#L2).

Next: [decisions and gotchas](05-decisions.md).
