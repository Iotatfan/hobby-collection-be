# Hobby Collection - Backend

## Stack
- Language: Go
- Framework: Gin
- ORM: GORM
- Database: PostgreSQL
- Architecture: Clean Architecture (entity → repository → service → handler)

## Project Structure
- `cmd/app/` — main entry point
- `cmd/migrate/` — database migration entry point
- `internal/collection/` — collection domain
  - `entity/` — structs and response types
  - `repository/` — DB queries via GORM
  - `service/` — business logic
  - `handler/` — HTTP handlers (Gin)

## Conventions
- Always pass `context.Context` as the first param in service/repo methods
- Use `response.go` structs for all API responses
- Errors should be returned, not panicked
- SQL filters use GORM's raw query with `?` placeholders

## DB Migration
- Migration script is located in `cmd/migrate/main.go`
- Run `go run cmd/migrate/main.go` to apply migrations
- Use proper logging for all migration steps

## Entity Layer
- Structs are defined in `internal/collection/entity/`
- `collection.go` contains main collection entity
- `request.go` contains request DTOs
- `response.go` contains API response structs
- `filter.go` contains filter parameters

## Repository Layer
- Methods defined in `internal/collection/repository/repository.go`
- Use GORM's `db.Where().Order().Limit().Offset().Find()` for queries
- Use `db.Create()`, `db.Save()`, `db.Delete()` for modifications
- Return `(entity.Collection, error)` or `([]entity.Collection, error)`

## Service Layer
- Methods defined in `internal/collection/service/service.go`
- Business logic goes here
- Map request DTOs to entity structs
- Call repository methods and handle errors
- Return appropriate response DTOs

## Handler Layer
- HTTP handlers in `internal/collection/handler/handler.go`
- Grouped under `/collections` route
- Use Gin's `c.BindJSON()`, `c.BindQuery()`
- Return JSON responses with `response.ApiResponse`
- Handle pagination using `limit`, `offset` query params

## Request/Response Mapping
- `request.go` contains DTOs that match API request format
- `response.go` contains DTOs that match API response format
- Handlers map between request DTOs, entities, and response DTOs
