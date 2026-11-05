# Redish Benchmark & Performance Evaluation

This document contains empirical performance benchmarks for `Redish`, evaluated across in-memory operations, concurrent multi-client network workloads, atomic transactions, and Multi-Part AOF persistence.

---

## 1. Test Environment & Hardware Specifications

- **CPU**: AMD Ryzen 7 7840HS (8 physical cores, 16 threads, up to 5.1 GHz)
- **RAM**: 32 GB DDR5
- **OS**: Linux (kernel 6.x, x86_64)
- **Go Version**: `go1.27.1 linux/amd64`
- **Network**: Loopback TCP (`127.0.0.1`)
- **Protocol**: Standard Redis RESP (REdis Serialization Protocol)
- **Value Payload**: 32 bytes

---

## 2. End-to-End Network Throughput & Latency

Evaluated using `cmd/benchmark` with persistent TCP connections and synchronous RESP request-response validation. Zero errors (0.00% failure rate) were recorded across all tests.

### 2.1 Concurrency Scaling (10, 50, and 100 Concurrent Clients)

#### A. 10 Concurrent Clients (50,000 requests per command)
| Command | Concurrency | Throughput (Req/sec) | Avg Latency | p50 (Median) | p95 Latency | p99 Latency | Error Rate |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `PING` | 10 | **332,800 req/s** | 29 µs | 27 µs | 44 µs | 58 µs | 0.00% |
| `SET` | 10 | **300,436 req/s** | 31 µs | 28 µs | 50 µs | 104 µs | 0.00% |
| `GET` | 10 | **295,836 req/s** | 32 µs | 29 µs | 50 µs | 89 µs | 0.00% |
| `INCR` | 10 | **305,257 req/s** | 31 µs | 28 µs | 49 µs | 93 µs | 0.00% |
| `RPUSH` | 10 | **295,717 req/s** | 31 µs | 28 µs | 49 µs | 86 µs | 0.00% |
| `LRANGE` | 10 | **246,023 req/s** | 38 µs | 33 µs | 63 µs | 126 µs | 0.00% |
| `MIXED (50/50)` | 10 | **299,030 req/s** | 31 µs | 28 µs | 53 µs | 89 µs | 0.00% |
| `TX (MULTI/EXEC)` | 10 | **80,924 tx/s** | 123 µs | 116 µs | 160 µs | 240 µs | 0.00% |

#### B. 50 Concurrent Clients (50,000 requests per command)
| Command | Concurrency | Throughput (Req/sec) | Avg Latency | p50 (Median) | p95 Latency | p99 Latency | Error Rate |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `PING` | 50 | **622,996 req/s** | 75 µs | 60 µs | 156 µs | 283 µs | 0.00% |
| `GET` | 50 | **588,515 req/s** | 79 µs | 62 µs | 166 µs | 330 µs | 0.00% |
| `INCR` | 50 | **590,746 req/s** | 78 µs | 64 µs | 159 µs | 274 µs | 0.00% |
| `RPUSH` | 50 | **589,832 req/s** | 79 µs | 62 µs | 166 µs | 312 µs | 0.00% |
| `SET` | 50 | **509,912 req/s** | 92 µs | 65 µs | 237 µs | 525 µs | 0.00% |
| `LRANGE` | 50 | **425,832 req/s** | 112 µs | 88 µs | 240 µs | 617 µs | 0.00% |
| `MIXED (50/50)` | 50 | **402,728 req/s** | 119 µs | 81 µs | 322 µs | 497 µs | 0.00% |
| `TX (MULTI/EXEC)` | 50 | **161,513 tx/s** | 303 µs | 258 µs | 557 µs | 1.05 ms | 0.00% |

#### C. 100 Concurrent Clients (100,000 requests per command)
| Command | Concurrency | Throughput (Req/sec) | Avg Latency | p50 (Median) | p95 Latency | p99 Latency | Error Rate |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `INCR` | 100 | **686,763 req/s** | 136 µs | 103 µs | 309 µs | 679 µs | 0.00% |
| `PING` | 100 | **682,785 req/s** | 137 µs | 100 µs | 318 µs | 723 µs | 0.00% |
| `GET` | 100 | **659,273 req/s** | 143 µs | 107 µs | 328 µs | 780 µs | 0.00% |
| `RPUSH` | 100 | **634,925 req/s** | 144 µs | 102 µs | 335 µs | 883 µs | 0.00% |
| `SET` | 100 | **575,984 req/s** | 163 µs | 106 µs | 508 µs | 977 µs | 0.00% |
| `LRANGE` | 100 | **464,598 req/s** | 206 µs | 142 µs | 642 µs | 1.48 ms | 0.00% |
| `MIXED (50/50)` | 100 | **389,681 req/s** | 247 µs | 89 µs | 815 µs | 1.31 ms | 0.00% |
| `TX (MULTI/EXEC)` | 100 | **175,570 tx/s** | 556 µs | 455 µs | 1.18 ms | 2.52 ms | 0.00% |

> *Note on `TX`: Each transaction executes a 4-step pipeline (`MULTI` -> `SET` -> `INCR` -> `EXEC`). At 175,570 transactions/second, the server processes over **702,000 transactional operations/second**.*

---

## 3. Multi-Part AOF Persistence Benchmarks

Under persistent disk logging (`-appendonly -appendfsync everysec`), write operations are logged to append-only files and synced periodically to disk.

**Configuration**: 50 concurrent clients, 50,000 requests per command.

| Command | Workload Type | Throughput (Req/sec) | Avg Latency | p50 (Median) | p99 Latency | Error Rate |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: |
| `MIXED` | 50% Read / 50% Persistent Write | **204,579 req/s** | 236 µs | 28 µs | 1.48 ms | 0.00% |
| `INCR` | Persistent Atomic Increment | **138,636 req/s** | 349 µs | 21 µs | 1.79 ms | 0.00% |
| `RPUSH` | Persistent List Append | **138,045 req/s** | 348 µs | 22 µs | 1.78 ms | 0.00% |
| `SET` | Persistent String Write | **134,312 req/s** | 360 µs | 22 µs | 1.81 ms | 0.00% |

---

## 4. In-Engine Store & Protocol Microbenchmarks

Evaluated via Go's standard benchmark tool (`go test -bench=. -benchmem -benchtime=2s`):

| Benchmark Name | Iterations | Latency (ns/op) | Memory Allocated | Allocs / Op |
| :--- | :---: | :---: | :---: | :---: |
| `BenchmarkStore_ParallelGet-16` | 73,909,028 | **34.21 ns/op** (~29.2M ops/s) | 15 B/op | 1 alloc |
| `BenchmarkStore_Incr-16` | 60,663,903 | **42.38 ns/op** (~23.6M ops/s) | 7 B/op | **0 allocs** |
| `BenchmarkStore_Rpush-16` | 49,632,014 | **59.91 ns/op** (~16.7M ops/s) | 99 B/op | **0 allocs** |
| `BenchmarkStore_Get-16` | 30,447,736 | **71.56 ns/op** (~14.0M ops/s) | 15 B/op | 1 alloc |
| `BenchmarkStore_ParallelMixed_80/20-16` | 12,764,175 | **187.80 ns/op** (~5.3M ops/s) | 28 B/op | 2 allocs |
| `BenchmarkStore_ParallelSet-16` | 11,470,762 | **211.00 ns/op** (~4.7M ops/s) | 80 B/op | 2 allocs |
| `BenchmarkStore_Set-16` | 6,499,010 | **410.40 ns/op** (~2.4M ops/s) | 156 B/op | 3 allocs |
| `BenchmarkStore_Lrange-16` | 4,060,248 | **540.90 ns/op** (~1.8M ops/s) | 2112 B/op | 5 allocs |
| `BenchmarkRESP_Formatter_SimpleString-16` | 132,762,116 | **18.04 ns/op** | 16 B/op | 1 alloc |
| `BenchmarkRESP_Formatter_ArrayCommand-16` | 14,534,307 | **171.50 ns/op** | 184 B/op | 7 allocs |
| `BenchmarkRESP_ToRESP-16` | 3,955,318 | **607.00 ns/op** | 416 B/op | 16 allocs |

---

## 5. How to Reproduce

To re-run the complete benchmark suite:

```bash
# 1. Run all microbenchmarks & network benchmarks automatically
./scripts/run_benchmarks.sh

# 2. Or run custom client load test against any running Redish instance
go run cmd/benchmark/main.go -port 6379 -c 50 -n 100000 -format markdown
```
