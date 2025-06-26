# Realtime iGaming Transaction Engine (Golang + PostgreSQL)

This project is a solution to demonstrate how an Entain iGaming platform can process game outcomes and keep player balances perfectly consistent even under load. The code is 100 % Golang, the ledger sits in PostgreSQL, and every design choice (from idempotent writes to JSON logs and Prometheus counters) is aimed at the realities of high‑throughput, money‑sensitive systems.

The system is also built to be compatible with automated evaluation tools. All endpoints return standard JSON responses with proper HTTP status codes, and three predefined users (1, 2, 3) ensure the system can be tested immediately after spinning up, no manual setup required.

What you can expect out of the box: I can spin it up with a single make up, hammer it with 25 RPS, and still observe:

* atomic balance changes (no double‑spend)

* duplicate‐safe payouts (thanks to transaction IDs)

* clear JSON logs for every request

* ready‑to‑scrape Prometheus metrics.

---

## Project Structure

```
.
├── Dockerfile                      # Docker setup for building the Go app image
├── Makefile                        # CLI shortcuts for build, run, test, etc.
├── README.md                       # Project documentation with setup, usage, and design notes
├── build
│   ├── docker-compose.yml         # Docker Compose config to run app + Postgres together
│   └── init.sql                   # SQL script to initialize DB schema and seed users
├── cmd
│   └── server
│       └── main.go                # Application entrypoint — initializes server, routes, logging
├── configs
│   └── config.go                  # Loads DB and env configs (host, port, etc.)
├── go.mod                          # Go module definition and dependencies
├── go.sum                          # Dependency version hashes (used by Go)
├── internal
│   ├── db
│   │   ├── migrate.go             # Loads and executes `init.sql` at startup
│   │   └── postgres.go           # Connects to Postgres and handles DB pooling
│   └── user
│       ├── handler.go            # HTTP handlers for /transaction and /balance
│       ├── model.go              # User and Transaction data models
│       └── service.go            # Business logic (balance updates, idempotency, etc.)
├── pkg
│   └── utils
│       ├── logging.go            # Structured logging setup using logrus
│       ├── middleware.go         # HTTP middleware (logging, recovery, etc.)
│       ├── ratelimiter.go        # Token bucket rate limiter middleware
│       ├── response.go           # Utility functions for standardized JSON responses
│       └── validate.go           # Request and header validation helpers
├── scripts
│   └── test.sh                   # Bash script to run API tests
└── test
    ├── load_test.go              # Load testing with simulated RPS
    ├── user_handler_test.go      # API-level unit tests for handlers
    └── user_test.go              # Tests for core user service logic

```

---

## How to Run

### Prerequisites

* Docker + Docker Compose
* `make` (recommended)

### Clone the Repo

```bash
git clone https://github.com/prakharsdev/entain-app.git
cd entain-app
```
### Run the service:

```bash
make up
```

This spins up Postgres and app container.
App runs on: http://localhost:8080

DB runs on: localhost:5432
---

## Testing

### Functional test script:

```bash
make test-script
```

✔️ Sends sample transaction and verifies balance

### Run all tests:

```bash
make test
```

✔️ Runs:

* Load test (25 RPS)
* Mux route test
* Full API test (transaction + balance check)

---

## Completed vs Requirements

| Feature                              | Status        |
| ------------------------------------ | --------------|
| `POST /user/{id}/transaction`        | Implemented   |
| `GET /user/{id}/balance`             | Implemented   |
| Validations (state, amount, headers) | Done          |
| Idempotency (transactionId)          | Done          |
| Balance cannot go negative           | Done          |
| Handles 20–30 RPS                    | Load-tested   |
| Docker + Compose support             | Fully working |
| Users 1, 2, 3 pre-created            | Done          |
| Docs on how to run/test              | (this README) |

---


## Features Implemented

### 1. **HTTP API Endpoints**

* `POST /user/{userId}/transaction` – Accepts transactions and updates user balance
* `GET /user/{userId}/balance` – Returns current balance (as JSON: { "userId": <uint64>, "balance": "<string with 2 decimals>" })

### 2. **Idempotency**

* Duplicate transaction IDs are rejected silently as "already processed"
* To test:

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: game" \
    -H "Content-Type: application/json" \
    -d '{"state":"win", "amount":"10.00", "transactionId":"dup_txn"}'
  # Repeat the same request
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: game" \
    -H "Content-Type: application/json" \
    -d '{"state":"win", "amount":"10.00", "transactionId":"dup_txn"}'
  ```

  ✔️ Second request will return: `{ "message": "Transaction already processed" }`

### 3. **Atomic Balance Updates with Negative Balance Protection**

* `lose` transactions check and block insufficient balance updates
* Amounts are parsed and stored as strings with up to 2 decimal places, matching the spec
* To test:

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: game" \
    -H "Content-Type: application/json" \
    -d '{"state":"lose", "amount":"100000", "transactionId":"big_loss"}'
  ```

  ✔️ Should return: `{ "error": "insufficient balance" }`

### 4. **Source-Type Header Validation**

* Only accepts `game`, `server`, or `payment` (case-insensitive)
* To test:

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: hacker" \
    -H "Content-Type: application/json" \
    -d '{"state":"win", "amount":"1.00", "transactionId":"bad_src"}'
  ```

  ✔️ Should return: `{ "error": "Missing or invalid Source-Type header" }`

### 5. **Predefined Users**

* Users `1`, `2`, and `3` are pre-seeded in the DB with balances as strings, and IDs stored as uint64
* To test:

  ```bash
  curl http://localhost:8080/user/1/balance
  ```

  Should return: `{ "userId": 1, "balance": "0.00" }`

### 6. **Rate Limiting**

* Each IP is limited using token-bucket algorithm (\~2 requests/sec, burst of 3)
* To test:
  Use a load test like `ab`, `wrk`, or hammer with curl in a loop
  After a few rapid requests, it should return HTTP 429 Too Many Requests

### 7. **Logging (Structured)**

* All logs use `logrus` in JSON format and output to stdout
* To test:

  ```bash
  docker-compose logs -f app
  ```

  Logs should appear like:

  ```json
  {"level":"info","method":"POST","path":"/user/1/transaction","duration":12}
  ```

### 8. **Prometheus Metrics**

* `/metrics` exposes default Go + Prometheus client metrics
* To test:

  ```bash
  curl http://localhost:8080/metrics
  ```

  You’ll see output like:

  ```
  go_gc_duration_seconds{quantile="0"} 0
  http_requests_total{code="200",method="get"} 5
  ```

### 9. **Panic Recovery**

* Gracefully handles panics and avoids crashing the server
* To test:
  Artificial panic can be injected in code and validated with logs

### 10. **Health Check Endpoint**

* `/health` returns app + DB status
* To test:

  ```bash
  curl http://localhost:8080/health
  ```

  Returns: `{ "status": "ok", "database": "connected" }`

---

## Design Highlights

* **Safe DB access** with row-level locking
* **SQL Migrations + Seeding** via code, no external migration tool
* **Test Coverage**: 3 integration tests
* **Makefile** to simplify common tasks
* **Clean architecture**: separates config, logic, utils, transport

---
## Design Decisions

### Language and Tech Stack

I chose **Golang** for its strong concurrency model, fast execution, and first-class support for building low-latency network services. For the database, I used **PostgreSQL** because it's battle-tested, supports strong consistency guarantees, and handles concurrent transactions well — a critical need for financial systems.

### API Structure

I kept the HTTP layer simple with only two endpoints (`/transaction` and `/balance`). This mirrors real-world iGaming APIs, where minimalism and clarity are essential for frontend or game-engine integrations.

### Idempotency & Transaction Logic

Idempotency was implemented using a unique constraint on the `transaction_id` and checking for duplicates before executing a DB write. This is crucial in gaming where duplicate payouts must be avoided at all costs due to retries or race conditions.

Balance updates are wrapped inside DB transactions with row-level locking using `SELECT FOR UPDATE`. This ensures atomicity and consistency even under concurrent loads — exactly what high-traffic systems need.

### Header and Payload Validations

I validated `Source-Type`, `state`, and `amount` to catch malformed or malicious inputs. Enforcing strict typing and enumerations avoids downstream issues in real money platforms where trust boundaries matter.

### Logging and Observability

I integrated `logrus` for structured logging in JSON format. This ensures compatibility with centralized logging stacks like ELK or Loki. Prometheus metrics were exposed from `/metrics` endpoint to allow scraping via a monitoring agent — giving visibility into system health, request counts, and latency.

### Error Handling

I implemented middleware to recover from panics gracefully and always return JSON-formatted errors with appropriate HTTP status codes. This helps the frontend and automated tools handle edge cases without crashing.

### Configuration

Database credentials and connection parameters are loaded via environment variables with sane defaults, which makes this system 12-factor compliant and easy to deploy across environments.

### Makefile + Docker + Compose

To ensure that the project can be spun up with zero configuration, I included a `Makefile` and Docker setup. `make up` brings up the full system with one command, and `make test` runs functional tests. This helps simulate CI/CD pipelines and improves DX.

### Security and Rate Limiting

To prevent abuse and brute-force attacks, I used token-bucket rate limiting per IP. While basic, it sets the foundation for applying more advanced auth or rate control mechanisms later (e.g., JWT, API keys).

### Predefined Users and Seed Data

Users 1, 2, and 3 are auto-seeded using SQL migrations. This makes the service testable immediately after container startup, which aligns with automated evaluation expectations.

> **Note on Migrations:**
> For simplicity and transparency, I have used a plain SQL script (`build/init.sql`) to handle schema creation and user seeding. This avoids dependency on external tools and ensures the app is immediately testable after startup.
> In production, I would migrate to a versioned tool like `golang-migrate` to handle evolving schemas cleanly across environments.

---


## Future Improvements

The current solution is already battle-ready for its intended scope, it handles transactions correctly, scales to 25+ RPS, and comes with strong observability, validation, and developer ergonomics. That said, there are a few realistic, high-leverage improvements I’d consider next:

### Vertical Improvements (Deeper capabilities in current system)

1. **Move to a Real Migration Tool**
   I’ve used SQL scripts embedded in Go for simplicity, but in a real deployment, I’d switch to a tool like [`golang-migrate`](https://github.com/golang-migrate/migrate) to handle versioned schema changes more cleanly across environments.

2. **Persist Metrics Externally**
   Right now, Prometheus metrics are exposed locally. I’d wire this up to a full Prometheus + Grafana stack (possibly via Docker) to visualize request rates, error counts, and latency trends over time.

3. **Test Coverage Expansion**
   I’ve focused primarily on functional and integration tests. I’d extend coverage by adding unit tests using `pgxmock` to test the service logic in isolation from the DB.

4. **Custom Error Codes + i18n-Ready Messages**
   Currently, error messages are developer-friendly. For production use, I’d build a proper error struct with codes and allow future support for internationalized responses (especially useful in gaming platforms across regions).

---

### Horizontal Improvements (Extending scope / capabilities)

1. **OpenAPI Documentation**
   I’d integrate Swagger UI using [`swaggo`](https://github.com/swaggo/swag) so future consumers (e.g., game clients or finance back offices) can easily discover and try the API.

2. **Authentication and Authorization**
   The service is public by design (for testability), but I’d plug in lightweight JWT-based auth middleware to restrict sensitive routes — especially balance lookups and payouts.

3. **Configurable Rate Limiting**
   Right now, rate limits are hardcoded. In production, I’d expose them via environment variables or a YAML config to allow tuning per environment (staging vs prod).

4. **Extended Transaction Support**
   The current state model is `win` / `lose`, which works well. But in the future, I’d generalize the engine to support additional transaction types like `refund`, `rollback`, or `bonus`, while keeping the same core integrity checks.

5. **Multi-Currency Support**
   While not immediately needed, real-world iGaming systems often deal with multiple currencies or tokens. I’d refactor the balance logic to be currency-aware (probably with a `currency` column and ISO validation).

---


## Clean Up

```bash
make down
```

