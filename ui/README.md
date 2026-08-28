# CoolDB Local Studio

This directory contains a local-only query workbench and a throwaway dashboard design prototype. It is intended for demos, not deployment.

## Run the complete demo

From the repository root:

```bash
make demo
```

This command builds `cool`, starts the gRPC server and opt-in HTTP demo bridge, installs UI dependencies when needed, and launches Next.js at [http://localhost:3000](http://localhost:3000).

The scratch database is stored at `.cooldb-demo/demo.cooldb` and is ignored by Git. Override it with `COOLDB_DEMO_DB` when needed.

## Dashboard variants

The root route contains three deliberately different prototypes:

- `?variant=A` — Command center: dense database explorer and operational workspace.
- `?variant=B` — Query notebook: a calmer, document-oriented exploration flow.
- `?variant=C` — Terminal wall: a keyboard-first, system-console presentation.

In development, use the floating arrow switcher or the keyboard arrow keys. The switcher is excluded from production builds.

## Run the processes separately

Start CoolDB with the demo bridge:

```bash
make build
./bin/cool server --db ./.cooldb-demo/demo.cooldb --http-port 3041
```

Then start the UI:

```bash
cd ui
npm install
npm run dev
```

The Next.js route handlers proxy `/api/health` and `/api/query` to `http://127.0.0.1:3041`. Set `COOLDB_DEMO_API_URL` to override that address.

## Safety

The workbench can execute mutating SQL against the selected database. The HTTP bridge is disabled by default in normal server runs and should not be exposed publicly. Use a scratch database for demos.
