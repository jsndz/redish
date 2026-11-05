# Redish

`Redish` is a high-performance, concurrent, lightweight Redis clone written in Go. It implements a core subset of the Redis RESP (REdis Serialization Protocol) protocol, support for basic key-value storage with expiration, list data structures, optimistic locking transactions, pub/sub channels, master-replica replication, and Multi-Part AOF persistence.

Unlike standard Redis, which is single-threaded, `Redish` handles multiple client connections concurrently using Go's goroutines and synchronized thread-safe data structures, achieving up to **686,000+ requests/sec** under multi-client network load.

---

## Performance Highlights

Tested under loopback TCP on **AMD Ryzen 7 7840HS** (16 threads, Linux 6.x, Go 1.27):

| Workload / Benchmark | Concurrency | Throughput | Median (p50) Latency | p99 Latency | Failure Rate |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **`INCR` (Atomic Counter)** | 100 clients | **686,763 req/s** | 103 µs | 679 µs | **0.00%** |
| **`GET` (Read Throughput)** | 100 clients | **659,273 req/s** | 107 µs | 780 µs | **0.00%** |
| **`RPUSH` (List Append)** | 100 clients | **634,925 req/s** | 102 µs | 883 µs | **0.00%** |
| **`SET` (Write Throughput)** | 100 clients | **575,984 req/s** | 106 µs | 977 µs | **0.00%** |
| **`LRANGE` (List Range Query)** | 100 clients | **464,598 req/s** | 142 µs | 1.48 ms | **0.00%** |
| **`TX` (`MULTI`/`EXEC` Pipeline)** | 100 clients | **175,570 tx/s** (*702k ops/s*) | 455 µs | 2.52 ms | **0.00%** |
| **Multi-Part AOF (`everysec`)** | 50 clients | **138,636 write/s** | 21 µs | 1.79 ms | **0.00%** |
| **In-Engine Store Read (`ParallelGet`)** | Engine Microbench | **~29.2M ops/s** (34.2 ns/op) | - | - | **1 alloc/op** |
| **In-Engine Atomic Increment (`Incr`)** | Engine Microbench | **~23.6M ops/s** (42.4 ns/op) | - | - | **0 allocs/op** |

> For comprehensive latency percentiles (p50, p90, p95, p99), memory allocations, and persistence breakdown, see the full [Benchmark Documentation](docs/BENCHMARKS.md).

---

## Features

### 1. Key-Value Store & Lists
- **Basic K-V**: Store strings and retrieve them with microsecond latency. Supports key expiration via `EX` (seconds) and `PX` (milliseconds).
- **Incr**: Atomically increment integer values with zero heap allocations.
- **List Support**: Push elements to the right of a list (`RPUSH`) and query slices (`LRANGE`).

### 2. Transactions & Optimistic Locking
- **Transactions**: Queue commands within atomic execution blocks using `MULTI`, execute them atomically with `EXEC`, or discard queued commands via `DISCARD`.
- **Optimistic Locking**: Monitor keys using `WATCH`. If any watched key is mutated by another client prior to `EXEC`, the transaction aborts (Check-And-Set / CAS semantics).

### 3. Pub/Sub Messaging
- **Publish-Subscribe**: Subscribe to channels using `SUBSCRIBE`, unsubscribe using `UNSUBSCRIBE`, and broadcast real-time messages via `PUBLISH`.

### 4. Master-Replica Replication
- **Handshake Sequence**: Replicas automatically handshake with the master server via `PING` -> `REPLCONF` (port) -> `REPLCONF` (capa) -> `PSYNC`.
- **Full Resynchronization**: Initiates full resync via dummy/mock RDB snapshots when replica states are empty.
- **Write Propagation**: Master propagates write commands to all connected active replicas.
- **Durable Write Acknowledgment**: The `WAIT` command pauses execution until the specified number of replicas acknowledge write offsets or timeout expires.

### 5. Multi-Part AOF Persistence
- **AOF (Append Only File)**: Durably logs state-modifying write commands to disk.
- **Multi-Part Structure**: Implements manifest-tracked base and incremental AOF logs.
- **Fsync Policy**: Configurable via `-appendfsync` (`always` or periodic `everysec`).

---

## Supported Commands

| Command | Arguments | Description |
| :--- | :--- | :--- |
| `PING` | | Returns `PONG`. |
| `ECHO` | `message` | Echoes back the message. |
| `SET` | `key` `value` `[EX seconds]` `[PX ms]` | Stores a string value, optionally setting an expiration. |
| `GET` | `key` | Retrieves the value of the key. |
| `INCR` | `key` | Atomically increments the numeric value of the key by 1. |
| `RPUSH` | `key` `value [value ...]` | Appends one or multiple values to a list key. |
| `LRANGE` | `key` `start` `stop` | Returns a range of elements from the list. |
| `MULTI` | | Marks the start of a transaction block. |
| `EXEC` | | Executes all queued commands in a transaction atomically. |
| `DISCARD` | | Flushes all queued commands in a transaction. |
| `WATCH` | `key [key ...]` | Monitors one or more keys for changes before a transaction. |
| `UNWATCH` | | Flushes all watched keys for the current client. |
| `SUBSCRIBE` | `channel [channel ...]` | Subscribes the client to specified channels. |
| `UNSUBSCRIBE`| `channel [channel ...]` | Unsubscribes the client from specified channels. |
| `PUBLISH` | `channel` `message` | Publishes a message to a channel. |
| `WAIT` | `numreplicas` `timeout` | Blocks until the specified number of replicas acknowledge write offset. |
| `CONFIG GET` | `parameter` | Retrieves configuration parameter (`dir`, `appendonly`, etc.). |

---

## Project Structure

```
.
├── cmd/
│   ├── main.go               # Server bootstrap and entrypoint
│   └── benchmark/main.go     # High-concurrency RESP benchmark load tester
├── internal/
│   ├── aof/                  # Multi-Part AOF manager, parser & state recovery
│   ├── channel/              # Pub/Sub channel multiplexing
│   ├── client/               # TCP client state, transaction queues & CAS flags
│   ├── commands/             # Modular command dispatchers & handlers
│   ├── config/               # Server configuration & CLI flag parsing
│   ├── core/                 # Replication protocol & offset tracking
│   ├── rdb/                  # RDB snapshot generator & parser
│   ├── server/               # TCP listener & concurrent connection handler
│   └── store/                # Thread-safe in-memory store & microbenchmarks
├── util/
│   ├── RESP.go               # Redis Serialization Protocol encoder & parser
│   └── resp_bench_test.go    # Protocol serialization/parsing microbenchmarks
├── docs/
│   ├── BENCHMARKS.md         # Full benchmark analysis & performance report
│   ├── AOF.md                # Multi-Part AOF architectural design
│   ├── optimisticLocking.md  # CAS transaction implementation details
│   ├── pubsub.md             # Pub/Sub documentation
│   └── replication.md        # Master-replica replication handshake & state
└── scripts/
    └── run_benchmarks.sh     # Automated benchmark execution suite
```

---

## Configuration Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-port` | `6379` | Server TCP listening port |
| `-dir` | `.` | Directory for database and snapshot files |
| `-appendonly` | `false` | Enable Multi-Part AOF persistence |
| `-appenddirname` | `appendonlydir` | Directory name for AOF files |
| `-appendfilename`| `appendonly.aof` | Master filename for AOF |
| `-appendfsync` | `everysec` | Disk sync policy (`always` or `everysec`) |
| `-replicaof` | `""` | Connects to master at `"host:port"` as a replica |

---

## Getting Started

### Prerequisites
- Go 1.25 or later

### Building and Running the Server

```bash
# Run server on default port 6379
go run cmd/main.go

# Run server with Multi-Part AOF persistence enabled
go run cmd/main.go -port 6379 -appendonly -appendfsync everysec

# Start a replica pointing to master
go run cmd/main.go -port 6381 -replicaof localhost:6379
```

### Interacting with Redish

Use standard `redis-cli` or raw TCP tools:

```bash
redis-cli -p 6379
127.0.0.1:6379> SET user:1001 "Alice" EX 60
OK
127.0.0.1:6379> GET user:1001
"Alice"
127.0.0.1:6379> INCR counter
(integer) 1
```

### Running Tests

```bash
# Run all unit and integration tests
go test ./...

# Run tests with race condition detector
go test -race ./...
```

### Running Benchmarks

```bash
# 1. Run the automated end-to-end benchmark suite
./scripts/run_benchmarks.sh

# 2. Run engine microbenchmarks with memory profiling
go test -bench=. -benchmem ./internal/store ./util

# 3. Run the custom concurrent load generator against any Redis/Redish instance
go run cmd/benchmark/main.go -port 6379 -c 50 -n 100000 -format markdown
```
