## Hobby Collection API

Go backend for the Hobby Collection site.

## Stack

- Go
- Gin
- GORM
- PostgreSQL
- Cloudinary

## Architecture

This project follows Clean Architecture principles with the following structure:

| Layer / Directory | Description |
| --- | --- |
| `entity/` | Core structs, request/response models, and filter parameters |
| `repository/` | Data access layer (database queries using GORM) |
| `service/` | Business logic layer (maps requests to entities and calls repositories) |
| `handler/` | HTTP delivery layer (Gin handlers, maps HTTP requests to service calls)

## Run locally

Create or update `config.yml` in the repo root, then run:

```powershell
go run .\cmd\app\main.go
```

The API listens on the port configured at `server.port`.

## Migrate database

```powershell
go run .\cmd\migrate\main.go
```

## Configuration

The app reads `config.yml` from the repo root. Environment variables can override nested config using `_`, for example `POSTGRES_HOST`.

Basic template:

```yaml
server:
  port: 8080

postgres:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "change-me"
  dbname: "change-me"

jwt:
  secret: "change-me"
  issuer: "change-me"
  audience: "change-me"
  required_role: "admin"

rate_limit:
  read_requests_per_minute: 120
  write_requests_per_minute: 20
  burst: 10

admin:
  password: "change-me"

cloudinary:
  name: "change-me"
  key: "change-me"
  secret: "change-me"
```

## Main endpoints

| Method | Endpoint | Auth |
| --- | --- | --- |
| `POST` | `/admin/token` | No |
| `GET` | `/collection` | No |
| `GET` | `/collection/:id` | No |
| `GET` | `/collection/filter` | No |
| `GET` | `/collection/drawer` | No |
| `GET` | `/collection/statistics` | No |
| `GET` | `/collection/shelves` | No |
| `POST` | `/create_collection` | JWT |
| `PATCH` | `/collection/:id` | JWT |
