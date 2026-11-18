# Architecture and Conventions

[Agent entry point](../agents.md)

## Repository map

| Location | Responsibility |
| --- | --- |
| `cmd/app/main.go` | Configuration, clients, middleware, domain registration, HTTP startup and graceful shutdown |
| `cmd/migrate/main.go` | GORM migrations and explicit schema/index statements |
| `internal/collection/` | Collection CRUD, catalogs, filters, drawers, shelves, statistics, and media orchestration |
| `internal/shortlink/` | Link creation, validation, expiry, threat checks, caching, and redirects |
| `internal/admin/` | Admin password verification and JWT issuance |
| `internal/router/` | Route registration and endpoint authentication middleware |
| `internal/common/` | Shared models, response envelope, error types, and HTTP error parsing |
| `internal/config/`, `internal/middleware/`, `internal/text/` | Configuration, HTTP middleware, and internal messages |
| `pkg/database/gorm/`, `pkg/cache/`, `pkg/storage/cloudinary/` | Infrastructure clients |

Collection implementation files are `collection_handler.go`, `collection_service.go`, and `collection_repository.go` in their respective subdirectories. Shortlink follows the same pattern with `shortlink_`. Constructors are connected in domain `module.go` files.

## Layer boundaries

- **Entities and DTOs:** collection models live in `internal/collection/entity/collection.go`; requests and filters are in `request.go`, and responses are in `response.go`. Shortlink has its own entity, request, and response files. Admin currently keeps token DTOs in its handler.
- **Handlers:** bind HTTP inputs, obtain request context, call services, and translate results into the established response contract. Keep persistence and business rules out of handlers.
- **Services:** apply business rules, map DTOs and domain data, and orchestrate repositories, media, threat detection, and caching. Extend the relevant service interface when adding operations.
- **Repositories:** use GORM `WithContext(ctx)` for database work and existing transaction patterns for related writes. Return the operation's existing entity or response DTO plus an error; not all queries return collection entities.
- **Composition:** construct dependencies in domain modules and register routes through `internal/router/`. Admin currently has no service or repository packages; do not invent them solely for uniformity.

## Persistence and errors

- Bind user-controlled SQL values, for example `Where("id = ?", id)`. Use the existing supported sort mapping instead of passing raw query parameters to `Order`.
- Preserve soft-delete handling in custom joins and raw queries, association updates, and transaction boundaries for collections, pictures, addons, and metadata tags.
- Propagate cancellation through database, cache, upload, and threat-check operations. Context-free constructors and pure or in-memory helpers are exceptions to the context-first I/O convention.
- Return errors and preserve causes where callers need `errors.Is`/`errors.As`. Use `common.ValError`, `DBError`, `JWTError`, or `ServiceError` consistently with existing boundaries.
- Inspect `internal/common/error_parser.go` before wrapping errors: status selection depends on concrete error types, and arbitrary wrapping is not automatically recognized there.
- Keep startup failures separate from request failures; entry points currently use fatal logging for initialization errors.

## Change checklist

Trace changes through DTOs, handler, service/repository interfaces, module wiring, and routes. For writes, inspect media handling, migrations, Redis invalidation, and the repository's in-memory filter cache. Follow the [API guide](api.md) for contracts and the [development guide](development.md) for validation.
