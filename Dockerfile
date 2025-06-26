# ----------- Build stage -----------
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the full project
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o entain-server ./cmd/server


# ----------- Test stage (with Go installed) -----------
FROM golang:1.24.4-alpine AS test-runner

WORKDIR /app

# Copy code and downloaded modules from builder
COPY --from=builder /app /app

# Set database connection string (can be overridden)
ENV DB_DSN=postgres://entain:entain@db:5432/entain_db?sslmode=disable

# Run unit tests (default command)
CMD ["go", "test", "./internal/user", "-v"]


# ----------- Minimal run stage -----------
FROM alpine:latest

WORKDIR /root/

# Copy binary only from builder
COPY --from=builder /app/entain-server .

EXPOSE 8080

# Run the application
ENTRYPOINT ["./entain-server"]
