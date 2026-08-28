#!/bin/sh

set -eu

project_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
grpc_port=${COOLDB_DEMO_GRPC_PORT:-3040}
http_port=${COOLDB_DEMO_HTTP_PORT:-3041}
ui_port=${COOLDB_DEMO_UI_PORT:-3000}
database_path=${COOLDB_DEMO_DB:-"$project_root/.cooldb-demo/demo.cooldb"}

mkdir -p "$(dirname "$database_path")"
make -C "$project_root" build

if [ ! -d "$project_root/ui/node_modules" ]; then
  npm --prefix "$project_root/ui" install
fi

"$project_root/bin/cool" server \
  --host 127.0.0.1 \
  --port "$grpc_port" \
  --http-port "$http_port" \
  --db "$database_path" &
server_pid=$!

cleanup() {
  kill "$server_pid" 2>/dev/null || true
  wait "$server_pid" 2>/dev/null || true
}
trap cleanup EXIT INT TERM

printf '\nCoolDB local demo\n'
printf '  Studio:  http://localhost:%s/?variant=A\n' "$ui_port"
printf '  API:     http://127.0.0.1:%s/api/health\n' "$http_port"
printf '  Database: %s\n\n' "$database_path"

COOLDB_DEMO_API_URL="http://127.0.0.1:$http_port" \
  npm --prefix "$project_root/ui" run dev -- --port "$ui_port"
