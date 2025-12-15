# quickserve

A lightweight, high-performance REST API server for user management in Go.

## Features

- Fast CRUD operations for users
- Thread-safe in-memory storage
- Zero external dependencies
- Built-in memory leak detection in tests

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /users | List all users |
| GET | /users/{id} | Get user by ID |
| POST | /users | Create new user |
| DELETE | /users/{id} | Delete user |
| GET | /health | Health check |

## Run

```bash
go run main.go
```

## Test with Leak Detection

```bash
go test -v ./...
```

## Example Requests

```bash
# Create user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# List users
curl http://localhost:8080/users

# Get user
curl http://localhost:8080/users/1

# Delete user
curl -X DELETE http://localhost:8080/users/1
```

## Heapcheck Integration

All tests use `guard.VerifyNone(t)` to detect goroutine and memory leaks:

```go
func TestHandleListUsers(t *testing.T) {
    defer guard.VerifyNone(t)
    // test code...
}
```
