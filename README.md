# Redish

`Redish` is a high-performance, concurrent, lightweight Redis clone written in Go. It implements a core subset of the Redis RESP (REdis Serialization Protocol) protocol, support for basic key-value storage with expiration, list data structures, optimistic locking transactions, pub/sub channels, master-replica replication, and Multi-Part AOF persistence.

Unlike standard Redis, which is single-threaded, `Redish` handles multiple clients concurrently using Go's goroutines and synchronized data structures.

---

## Features

### 1. Key-Value Store & Lists
- **Basic K-V**: Store strings and retrieve them with high speed. Supports expiration via `EX` (seconds) and `PX` (milliseconds).
- **Incr**: Atomically increment integer values.
- **List Support**: Push elements to the right of a list and query list ranges.

### 2. Transactions & Optimistic Locking
- **Transactions**: Queue commands using `MULTI`, execute them atomically with `EXEC`, or clear them with `DISCARD`.
- **Optimistic Locking**: Monitor keys using `WATCH`. If any watched key is modified by another client before `EXEC`, the transaction fails, preventing race conditions.

### 3. Pub/Sub Messaging
- **Publish-Subscribe**: Subscribe to channels using `SUBSCRIBE`, unsubscribe using `UNSUBSCRIBE`, and broadcast messages via `PUBLISH`.

### 4. Master-Replica Replication
- **Handshake Sequence**: Replicas automatically handshake with the master server via a `PING` -> `REPLCONF` (port) -> `REPLCONF` (capa) -> `PSYNC` protocol.
- **Full Resynchronization**: Initiates full resync via dummy/mock RDB snapshots when replica states are empty.
- **Write Propagation**: Master propagates write commands down to all connected replicas.
- **Durable Write Acknowledgment**: The `WAIT` command allows masters to pause and wait for write ACKs from replicas up to a specified offset and timeout, improving write durability.

### 5. Multi-Part AOF Persistence
- **AOF (Append Only File)**: Logs write commands to disk.
- **Multi-Part Structure**: Employs a manifest file tracking base and incremental AOF logs.
- **Fsync Policy**: Supports config option `always` or `everysec` (default) for syncing data writes to disk.

---

## Project Structure

The project code is organized as follows:

*   [cmd/main.go](file:///home/jaison/code/projects/redish/cmd/main.go) - Main entrypoint to set up and boot the server.
*   [internal/server/server.go](file:///home/jaison/code/projects/redish/internal/server/server.go) - Core TCP connection handler, replica handshaking, command dispatcher loop.
*   [internal/store/store.go](file:///home/jaison/code/projects/redish/internal/store/store.go) - Thread-safe in-memory key-value and list storage with support for TTL timers and watched keys.
*   [internal/aof/aof.go](file:///home/jaison/code/projects/redish/internal/aof/aof.go) - Append-Only File (AOF) manager implementing multi-part file parsing, manifest reading, and database state restoration.
*   [internal/commands/](file:///home/jaison/code/projects/redish/internal/commands) - Subdirectories for each command handler (e.g., set, get, ping, wait, multi, watch, etc.).
*   [internal/client/client.go](file:///home/jaison/code/projects/redish/internal/client/client.go) - Client connection metadata, transaction queues, and client-specific states.
*   [internal/config/config.go](file:///home/jaison/code/projects/redish/internal/config/config.go) - Command-line configuration parsing.
*   [internal/rdb/rdb.go](file:///home/jaison/code/projects/redish/internal/rdb/rdb.go) - Mock implementation of Redis Database (RDB) loading and generation.
*   [util/RESP.go](file:///home/jaison/code/projects/redish/util/RESP.go) - RESP protocol parser and formatter.

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
| `EXEC` | | Executes all queued commands in a transaction. |
| `DISCARD` | | Flushes all queued commands in a transaction. |
| `WATCH` | `key [key ...]` | Monitors one or more keys for changes before a transaction. |
| `UNWATCH` | | Flushes all watched keys for the current client. |
| `SUBSCRIBE` | `channel [channel ...]` | Subscribes the client to specified channels. |
| `UNSUBSCRIBE`| `channel [channel ...]` | Unsubscribes the client from specified channels. |
| `PUBLISH` | `channel` `message` | Publishes a message to a channel. |
| `WAIT` | `numreplicas` `timeout` | Blocks the client until the specified number of replicas acknowledge the write. |
| `CONFIG GET` | `parameter` | Gets configuration parameter value (`dir`, `appendonly`, etc.). |

---

## Configuration Flags

When starting the server, you can supply various command-line configuration flags:

```bash
# General Flags
-port             # Server port (default: 6379)
-dir              # Directory for database files (default: ".")

# Persistence Flags
-appendonly       # Enable Multi-Part AOF persistence (default: false)
-appenddirname    # Name of directory for AOF files (default: "appendonlydir")
-appendfilename   # Master filename for AOF (default: "appendonly.aof")
-appendfsync      # Fsync policy: 'always' or 'everysec' (default: "everysec")

# Replication Flags
-replicaof        # Connects to master at "host:port" to run as a replica
```

---

## Getting Started

### Prerequisites
- Go 1.25.4 or later

### Building and Running the Server
Run the following to compile and run the `Redish` server:

```bash
# Run server on default port 6379
go run cmd/main.go

# Run server on port 6380
go run cmd/main.go -port 6380

# Start a replica server pointing to master
go run cmd/main.go -port 6381 -replicaof localhost:6379
```

### Running Tests
To run all package tests:
```bash
go test ./...
```

### Connecting to Redish
You can interact with `Redish` using any standard Redis client (like `redis-cli`) or raw TCP tools:

```bash
redis-cli -p 6379
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET name antigravity
OK
127.0.0.1:6379> GET name
"antigravity"
```
