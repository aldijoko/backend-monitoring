#!/usr/bin/env bash
# Jalankan mediamtx + backend API bareng, matikan keduanya bareng dengan Ctrl+C.
set -euo pipefail

cd "$(dirname "$0")/.."

if ! docker ps --format '{{.Names}}' | grep -q '^bnpt-cpma-postgres$'; then
	echo "!! Container bnpt-cpma-postgres tidak jalan. Jalankan dulu sebelum lanjut." >&2
	exit 1
fi

if ! command -v mediamtx >/dev/null 2>&1; then
	echo "!! mediamtx tidak ditemukan. Install: brew install mediamtx ffmpeg" >&2
	exit 1
fi

pids=()
cleanup() {
	echo
	echo "Menghentikan proses..."
	for pid in "${pids[@]}"; do
		kill "$pid" 2>/dev/null || true
	done
	wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

mediamtx mediamtx/mediamtx.yml 2>&1 | sed -e 's/^/[mediamtx] /' &
pids+=("$!")

sleep 1

go run cmd/api/main.go 2>&1 | sed -e 's/^/[api]      /' &
pids+=("$!")

wait
