# go-redis

go-redis is a Go-based Redis implementation based on RESP (Redis Serialization Protocol) parsing. It is designed to provide a Redis-like experience with key features, including support for `string keys`, `graceful shutdown`, `AOF (Append-Only File) persistence`.

## Features

- `RESP Protocol Parsing`: Implements RESP protocol to allow Redis-like communication with clients.
- `String Keys`: Supports Redis-like string key storage and retrieval.
- `Graceful Shutdown`: Handles proper shutdown of Redis instances with in-progress requests managed.
- `AOF (Append-Only File)`: Implements AOF persistence for durability, logging all write operations.
  1. `Command Logging`: Every write operation is appended to the AOF file in RESP format.
  2. `Fsync Policy`:
     * **always**: Flushes data to disk after every command.
     * **everysec**: Flushes data every second asynchronously.
     * **no**: Relies on the operating system to decide when to flush.
  3. `AOF Buffering`: Commands are first written to an internal buffer before being flushed to the file.
  4. `AOF Rewrite Mechanism`: AOF rewrite is triggered under the following conditions:
     * The AOF file size exceeds a configured threshold (auto-aof-rewrite-min-size).
     * The file has grown by a defined percentage since the last rewrite (auto-aof-rewrite-percentage).
     * Manually triggered by the user (BGREWRITEAOF command).

## TODO

- [ ] `AOF Rewrite Incremental Fsync`: Implement incremental fsync while AOF rewrite.
- [x] `Authentication`: Support for authentication with password-based access control.
- [ ] `System Info Command`: Add the `system info` command to provide information about Redis server and its configuration.
- [ ] `RDB (Redis Database Persistence)`: Implement RDB persistence for saving snapshots of the database.
- [ ] `Redis Cluster`: Implement Redis Cluster to manage sharded data across multiple Redis nodes.
- [ ] `Redis Sentinel`: Implement Redis Sentinel for automatic fail-over and high availability.
- [ ] `Unit Tests`: Write unit tests to ensure correctness and reliability.
- [ ] `GitHub Actions`: Set up GitHub Actions for continuous integration and deployment.
- [ ] `All Redis Data Structures`: Implement all Redis data types, including Lists, Sets, Hashes, Sorted Sets, etc.
- [ ] `Pub/Sub Mechanism`: Implement the Publish/Subscribe (Pub/Sub) messaging system.

## Getting Started

1. `Clone the repository`:
    ```bash
    git clone https://github.com/POABOB/go-redis.git
    ```
2. `Install dependencies`: Follow the Go installation instructions if you haven’t set up Go yet: [Go Install](https://go.dev/doc/install).
3. `Run the server`: Once cloned, you can build and run the Redis server implementation with the following:
    ```bash
    go run main.go
    ```
4. `Client interactions`: You can interact with the server using a Redis client, using the RESP protocol to send commands to the server.

## Contributing
Contributions are welcome! If you want to help with the development of Redis Sentinel, Redis Cluster, or any other features, feel free to fork the repository, create a new branch, and submit a pull request.

## License
Distributed under the [MIT License](https://github.com/POABOB/go-redis/blob/main/LICENSE).