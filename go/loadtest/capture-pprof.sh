#!/usr/bin/env sh
set -eu

PPROF_BASE_URL="${PPROF_BASE_URL:-http://127.0.0.1:6060}"
CPU_SECONDS="${CPU_SECONDS:-30}"
OUTPUT_DIRECTORY="${OUTPUT_DIRECTORY:-$(dirname "$0")/results}"

mkdir -p "$OUTPUT_DIRECTORY"
go tool pprof -proto -seconds "$CPU_SECONDS" -output "$OUTPUT_DIRECTORY/cpu.pb.gz" \
  "$PPROF_BASE_URL/debug/pprof/profile"
go tool pprof -proto -output "$OUTPUT_DIRECTORY/heap.pb.gz" \
  "$PPROF_BASE_URL/debug/pprof/heap"
curl --fail --silent --show-error \
  "$PPROF_BASE_URL/debug/pprof/goroutine?debug=1" \
  --output "$OUTPUT_DIRECTORY/goroutine.txt"

echo "Profiles written to $OUTPUT_DIRECTORY"
