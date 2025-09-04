# 🏆 Leaderboard API

## Overview

The application is built on **The Twelve-Factor App** rules:

- **One code base** – GitHub
- **Clearly declared and isolated dependencies** – `go.mod`
- **Configuration** must be located in environment variables – `.env`
- **Strict separation of build, release, and execution** – CI/CD pipeline (future)
- **Stateless processes** – we store data in constant storage and update it (**Redis**, **PostgreSQL**)
- **Port binding** – the built-in web server runs on the specific port from the environment variable
- **Concurrency** – processes can be split into separate microservices in the future for high-load spots
- **Disposability** – supports graceful shutdown through a single `Context`
- **Logs** – currently using **Zap Logger**, future: **ELK Stack** (Elasticsearch + Logstash + Kibana)
- **Admin processes** – QA/Dev/Prod environments must be as similar as possible (future)

---

## DDD Architectural Structure

The application uses **DDD (Domain Driven Design)** architecture with fully separated layers:

- **Interface** – REST, controllers, middlewares, HTTP request/response DTOs
- **Application** – use cases, calls domain objects and repositories through interfaces
- **Domain** – domain entities (currently no complex business rules)
- **Infrastructure** – repository implementations, DB models, storages (Redis, SQL), gRPC/HTTP clients

**Request flow:**  
`client → interface → application → domain → infrastructure → DB`

**Dependencies** are directed inward: outer layers depend on inner ones, not vice versa.

---

## Concurrency Patterns

The application uses:
- **Worker Pool** – for parallel processing of events
- **Fan-in** – to combine results from multiple goroutines into a single output channel

---

## API Specifications

External contracts follow **OpenAPI standards**:  
`internal/interface/api/rest/api-specs/openapi/leaderboardapi/openapi.yaml`

All possible cURL requests are located here and can be run directly from your IDE (tested in GoLand):  
`internal/interface/api/rest/api-specs/leaderboardapi.http`

---

## Tests

- Pattern: **TableDrivenTests**
- Library: [`testify`](https://github.com/stretchr/testify)

Run from the root project directory to see code coverage:

```bash
$ go test ./... -cover
$ go test ./... -coverprofile=coverage.out
$ go tool cover -html=coverage.out
```

---

## Ops

Infrastructure endpoints for metrics and health checks:

-- `http://localhost:8080/metrics`:
 * "leaderboard_ingest_events_processed_total{result="**accepted**"}" - accepted Event counter label
* "leaderboard_ingest_events_processed_total{result="**duplicate**"}" - duplicate Event counter label

-- `http://localhost:8080/healthz`

---

## Application Initialization Steps

1. Create application
2. Get configuration
3. Init logs, clients, DBs, etc.
4. Run application including all parallel processes:
    - HTTP server
    - Cache `BackupWorker`
    - `ScorerPool` of workers for asynchronous event processing
    - `LeaderboardWorker` to update leaderboard from processed events
5. On `SIGURG` signal or context cancel, gracefully shut down the application

---

## Fault Tolerance and Load Optimization (future)

We will use **Replication** for load optimization, for example: **1 master, 2 replicas**.

- **Master** can store/update data in **PostgreSQL** and **Redis**
- **Replicas** can only read from Redis and keep the data in memory

If the application grows to a very large amount of Leaders data, we can use **Sharding** and even apply **Replication** for each shard.
