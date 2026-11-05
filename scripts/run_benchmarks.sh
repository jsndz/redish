#!/usr/bin/env bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$DIR/bin"
DOCS_DIR="$DIR/docs"
mkdir -p "$BIN_DIR" "$DOCS_DIR"

echo "=== 1. Building Server and Benchmark CLI ==="
go build -o "$BIN_DIR/redish" "$DIR/cmd/main.go"
go build -o "$BIN_DIR/redish-bench" "$DIR/cmd/benchmark/main.go"

echo "=== 2. Running In-Engine Store & RESP Microbenchmarks ==="
MICRO_BENCH_FILE="$DOCS_DIR/microbench_raw.txt"
go test -bench=. -benchmem -benchtime=2s "$DIR/internal/store" "$DIR/util" | tee "$MICRO_BENCH_FILE"

PORT=6399
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    rm -rf /tmp/redish_bench_aof_*
}
trap cleanup EXIT

echo "=== 3. Running End-to-End In-Memory Benchmarks ==="
"$BIN_DIR/redish" -port "$PORT" &
SERVER_PID=$!
sleep 1

echo "--- 3.1 Concurrency: 10 Clients, 50,000 Requests/cmd ---"
"$BIN_DIR/redish-bench" -port "$PORT" -c 10 -n 50000 -format markdown > "$DOCS_DIR/bench_c10.md"
cat "$DOCS_DIR/bench_c10.md"

echo "--- 3.2 Concurrency: 50 Clients, 50,000 Requests/cmd ---"
"$BIN_DIR/redish-bench" -port "$PORT" -c 50 -n 50000 -format markdown > "$DOCS_DIR/bench_c50.md"
cat "$DOCS_DIR/bench_c50.md"

echo "--- 3.3 Concurrency: 100 Clients, 100,000 Requests/cmd ---"
"$BIN_DIR/redish-bench" -port "$PORT" -c 100 -n 100000 -format markdown > "$DOCS_DIR/bench_c100.md"
cat "$DOCS_DIR/bench_c100.md"

kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""
sleep 1

echo "=== 4. Running Multi-Part AOF Persistence Benchmarks ==="
AOF_DIR="/tmp/redish_bench_aof_$$"
mkdir -p "$AOF_DIR"
"$BIN_DIR/redish" -port "$PORT" -dir "$AOF_DIR" -appendonly -appendfsync everysec &
SERVER_PID=$!
sleep 1

echo "--- 4.1 AOF Persistence (everysec): 50 Clients, SET & INCR ---"
"$BIN_DIR/redish-bench" -port "$PORT" -c 50 -n 50000 -t "SET,INCR,RPUSH,MIXED" -format markdown > "$DOCS_DIR/bench_aof.md"
cat "$DOCS_DIR/bench_aof.md"

echo "=== All Benchmarks Completed Successfully! ==="
