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
This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) for a clean and scalable Go codebase.

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
## Database Schema

This service uses two tables in PostgreSQL: `users` and `transactions`.

### Table Summary

#### `users`

| Column    | Type          | Description                 |
| --------- | ------------- | --------------------------- |
| `id`      | BIGINT        | Primary key (user ID)       |
| `balance` | NUMERIC(12,2) | User balance (default 0.00) |

#### `transactions`

| Column           | Type          | Description                      |
| ---------------- | ------------- | -------------------------------- |
| `transaction_id` | TEXT          | Primary key, ensures idempotency |
| `user_id`        | BIGINT        | Foreign key → users.id           |
| `amount`         | NUMERIC(12,2) | Amount (max 2 decimal places)    |
| `state`          | TEXT          | 'win' or 'lose'                  |
| `source_type`    | TEXT          | 'game', 'server', or 'payment'   |
| `created_at`     | TIMESTAMP     | Defaults to current timestamp    |

---

### ERD

assets/ERD.png

---

## How to Run

### Prerequisites

* Docker + Docker Compose
* `make` (recommended)
* WSL (if using Windows, tested on WSL2 + Ubuntu)

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

Sends sample transaction and verifies balance

### Run all tests:

```bash
make test
```

Runs:

* Load test (25 RPS)
* Mux route test
* Full API test (transaction + balance check)

---

## Alternative: Run with plain Docker Compose

If you're not using `make`, you can still start everything with Docker Compose:

```bash
docker-compose -f build/docker-compose.yml up -d --build
```

To shut it down and remove volumes:

```bash
docker-compose -f build/docker-compose.yml down -v
```

---

## Makefile Commands

| Command            | Description                           |
| ------------------ | ------------------------------------- |
| `make up`          | Build and start all containers        |
| `make down`        | Stop and remove all containers        |
| `make rebuild-app` | Rebuild only the app container        |
| `make logs`        | Tail logs from the app container      |
| `make clean`       | Remove generated Go binary            |
| `make test`        | Run all Go tests in `/test` directory |
| `make test-script` | Run bash-based API smoke test         |

---

## Coverage Summary of All Requirements

| Requirement from the Test Task                                | Covered in Features Section?       | Notes and Details                                                              |
| ------------------------------------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------ |
| `POST /user/{userId}/transaction` route                       | Yes – Feature 1                    | Handles transactions and updates user balances with sample requests            |
| `GET /user/{userId}/balance` route                            | Yes – Features 1, 5                | Returns balance as string with 2 decimal places in expected JSON format        |
| Validate `Source-Type` header                                 | Yes – Feature 4                    | Accepts only `game`, `server`, or `payment` (case-insensitive)                 |
| Accept and validate `state`, `amount`, `transactionId` fields | Yes – Feature 3                    | Validates allowed values and ensures format correctness                        |
| Only allow `win`, `lose` states                               | Yes – Feature 3                    | Uses `IsValidState()` for strict state validation                              |
| Ensure each `transactionId` is processed only once            | Yes – Feature 2                    | Fully idempotent; duplicate transactions return "already processed"            |
| Prevent user balance from going negative                      | Yes – Feature 3                    | `lose` requests fail gracefully if balance is insufficient                     |
| Predefined users 1, 2, 3 with valid IDs                       | Yes – Feature 5                    | Users are seeded in DB; validated with curl and unit tests                     |
| Ready to run with Docker Compose (no extra config)            | Yes – "How to Run" section         | Works out of the box using `docker-compose up` or `make up`                    |
| Testable via automation tools                                 | Yes – Entire project design        | Uses JSON responses, HTTP codes, and stable routes for automated validation    |
| Can handle 20–30 RPS                                          | Yes – Feature 6                    | Load tested via `load_test.go`; rate limits tested at 100 RPS                  |
| Balance returned as string (2 decimal places)                 | Yes – Features 1, 3                | Uses `strconv.FormatFloat(..., 2)` for consistent precision                    |
| `amount` field is string and limited to 2 decimal places      | Yes – Feature 3                    | Enforced via `IsValidAmountFormat()` logic in validation                       |
| Proper HTTP status codes used                                 | Yes – Throughout                   | 200 OK for success, 400+ for validation and server errors                      |
| Structured logging in JSON                                    | Yes – Feature 7                    | Uses `logrus`; logs include duration, method, status, path, etc.               |
| Prometheus metrics exposed                                    | Yes – Feature 8                    | `/metrics` endpoint enabled and tested                                         |
| Health check endpoint                                         | Yes – Feature 10                   | `/health` returns both app and DB status in JSON                               |
| Graceful panic recovery                                       | Yes – Feature 9                    | Middleware catches panics and logs them instead of crashing the server         |
| Docker + Makefile automation                                  | Yes – Covered in Makefile section  | `make up`, `make down`, `make test` simplify running, testing, and debugging   |


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

  Second request will return: `{ "message": "Transaction already processed" }`

### 3. **Atomic Balance Updates with Negative Balance Protection**

* `lose` transactions check and block insufficient balance updates
* Amounts are parsed and stored as strings with up to 2 decimal places, matching the spec
* Transactions with more than 2 decimal places are rejected to ensure precision integrity
* To test:

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: game" \
    -H "Content-Type: application/json" \
    -d '{"state":"lose", "amount":"100000", "transactionId":"big_loss"}'
  ```

  Should return: `{ "error": "insufficient balance" }`

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: game" \
    -H "Content-Type: application/json" \
    -d '{"state":"win", "amount":"10.123", "transactionId":"too_precise"}'
  ```

  Should return: `{ "error": "Amount must have at most 2 decimal places" }`

### 4. **Source-Type Header Validation**

* Only accepts `game`, `server`, or `payment` (case-insensitive)
* To test:

  ```bash
  curl -X POST http://localhost:8080/user/1/transaction \
    -H "Source-Type: hacker" \
    -H "Content-Type: application/json" \
    -d '{"state":"win", "amount":"1.00", "transactionId":"bad_src"}'
  ```

  Should return: `{ "error": "Missing or invalid Source-Type header" }`

### 5. **Predefined Users**

* Users `1`, `2`, and `3` are pre-seeded in the DB with balances as strings, and IDs stored as uint64
* To test:

  ```bash
  curl http://localhost:8080/user/1/balance
  ```

  Should return: `{ "userId": 1, "balance": "0.00" }`

### 6. **Rate Limiting**

* To simulate high load and trigger rate limiting, modify rps = 100 in load_test.go, then run:
```bash
RATE_LIMIT_RPS=2 RATE_LIMIT_BURST=3 make test TEST_ARGS="-count=1"
```

You’ll observe:

* 429 Too Many Requests for excess calls
* Output includes success vs rate-limited count

Example:

```
--- FAIL: TestLoadTransactionEndpoint
    load_test.go:64: Processed 100 requests in 76.57ms
    load_test.go:65: Success: 62, RateLimited: 38, Other Errors: 0
```

### 7. **Logging (Structured)**

* All logs use `logrus` in JSON format and output to stdout
* To test:

  ```bash
  make logs 
  ```

  Logs should appear like:

  ```json
  {"duration":1,"level":"info","method":"GET","msg":"Handled request","path":"/user/1/balance","time":"2025-06-27T00:32:19Z"}
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

### Data Precision and Storage

* All **balances and amounts are stored as strings** in the database (e.g., `"9.25"` instead of float) to exactly match the expected response format and avoid floating-point inaccuracies. This also aligns with the spec which expects amounts as string with 2 decimal places.

* During validation, **amounts with more than 2 decimal places are rejected**. This ensures strict adherence to the defined precision limit and avoids rounding surprises in financial computations.

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

