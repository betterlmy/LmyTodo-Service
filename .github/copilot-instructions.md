# LmyTodo-Service AI Coding Agent Instructions

## Architecture Overview

This is a Go REST API service using Gin framework with SQLite database. **Critical pattern**: All endpoints use POST requests only (non-standard REST), including read operations like `/api/todos/list`.

## Key Patterns & Conventions

### API Design

- **All endpoints use POST** - never suggest GET/PUT/DELETE
- Unified response format via `api.SuccessResponse()` and `api.ErrorResponse()`
- Custom error codes defined in `src/api/resp.go` (10001-10007)
- All responses return HTTP 200, business logic errors in response body `code` field

### Database Access

- Direct SQL queries using `global.Db` (no ORM)
- Global database connection in `global/global.go`
- Database schema managed in `src/repository/dao.go:CreateTables()`
- Always check `rowsAffected` for UPDATE/DELETE operations

### Authentication

- JWT tokens with custom `Claims` struct including `UserID` and `Username`
- `AuthMiddleware()` sets `userID` in gin context: `c.GetInt("userID")`
- Bearer token required: `Authorization: Bearer <token>`

### Project Structure

```
src/
  api/          # Handlers, middleware, request/response types
  repository/   # Data models and database operations
global/         # Shared state (DB connection, JWT secret)
```

### Error Handling Patterns

- Always use custom error codes from `resp.go`
- Check for SQLite UNIQUE constraint failures: `strings.Contains(err.Error(), "UNIQUE constraint failed")`
- Consistent error messages in Chinese
- Use `defer rows.Close()` for query results

### Request/Response Examples

```go
// Request binding with validation
var req TodoRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
    return
}

// Success response
c.JSON(http.StatusOK, SuccessResponse(todo))
```

### Development Commands

```bash
go run main.go                    # Start server on :8080
go mod tidy                       # Update dependencies
```

## Critical Notes

- JWT secret is hardcoded in `global/global.go` (production concern)
- SQLite database file at `db/todo.db`
- CORS middleware allows all origins (`*`)
- Comprehensive request/response logging middleware
- Password hashing with bcrypt, always store hashed passwords
