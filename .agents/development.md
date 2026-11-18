# Development and Validation

[Agent entry point](../agents.md)

## Configuration and dependencies

Run commands from the repository root. `go.mod` declares Go 1.25.0; the current Docker builder uses `golang:1.26-alpine`. Check these files when changing toolchains rather than silently aligning versions during unrelated work.

For initial setup, copy `config.example.yml` to `config.yml` only if local configuration does not already exist, then supply local values. Treat `config.yml` and `.env` as private; do not print credentials or copy their contents into documentation.

`internal/config/config.go` defines the schema. Startup reads root `config.yml` with Viper; if missing, it creates a zero-value file that still needs valid settings. Viper maps nested environment names using `_` in place of `.`, such as `POSTGRES_HOST`. Keep overridden keys present in YAML and verify effective values when changing loading; do not assume arbitrary environment-only keys are unmarshaled. Local `go run` does not load `.env` itself.

| Section / variable | Use |
| --- | --- |
| `server.port` | HTTP listen port |
| `postgres` | Host, port, user, password, and database name (`dbname`) |
| `redis` | Host, port, password, database index (`database`, zero if omitted) |
| `cloudinary` | Cloud name, API key, and secret for media operations |
| `jwt` | Secret, issuer, audience, required role, and access lifetime |
| `admin.password` | Token-issuance password |
| `rate_limit` | Read/write requests per minute and burst allowance |
| `shortlink.fe_url` | Frontend base URL for failed-resolution redirects |
| `WEBRISK_API_KEY` | Direct environment lookup enabling Google Web Risk |

Validation requires a JWT secret of at least 32 characters, nonempty issuer/audience/required role, and an admin password of at least 16 characters. This also applies to migrations because they use the same loader.

PostgreSQL supplies persistence, Redis caching, and Cloudinary media storage. With `WEBRISK_API_KEY` set, shortlink initialization creates a Web Risk client. Without it, threat detection is disabled and the nil-client detector treats destinations as non-malicious. Do not imply threat screening is always enabled.

## Run and migrate

```powershell
go run ./cmd/app
```

The API listens on `server.port`. The local app entry point does not migrate automatically. To migrate the configured database:

```powershell
go run ./cmd/migrate
```

Inspect the target configuration and script before executing this database-changing command. It runs GORM `AutoMigrate` for collection/catalog/media and shortlink models, followed by explicit SQL, including removal of the legacy scales `collection_type_id` column and index. Preserve migration-step logging and error checks. Do not run migrations merely to validate documentation.

`docker-compose.yml` runs Redis and the API, loads `.env`, and mounts `config.yml`. PostgreSQL is commented out and must be supplied separately. Configure hosts reachable from inside the API container. The Docker image builds both binaries and runs migration before starting the API, so starting it can change the configured schema.

## Cache behavior

Collection services cache filter, list, detail, statistics, and shelf responses in Redis. Current TTLs are 12 hours for filters, 30 minutes for lists, and one hour for detail, statistics, and shelves. List keys depend on filters; shelf keys distinguish category, limit, offset, and effective category filters from previews. Release-type ID sets are sorted and deduplicated for cache identity without mutating requests; filtered variants use the existing shelf wildcard invalidation.

The collection repository also maintains a mutex-protected in-memory filter cache with a five-minute TTL. Review both layers when changing writes or catalogs. Existing collection invalidation covers detail, statistics, preview shelves, paged shelves, list variants, and the in-memory filter cache; catalog creation has its own filter invalidation path. Preserve cache-error fallback behavior in the affected method rather than assuming every path handles nil caches identically.

Shortlink cache entries include destination and expiry. Valid hits still undergo destination validation and configured threat checks. TTL tracks remaining lifetime; malicious-link detection removes cached entries. Use existing cache interfaces and stubs for focused service tests.

## Validation commands

For Go behavior changes, run focused package tests first, then the ordinary suite and static checks:

```powershell
go test ./...
go vet ./...
```

For concurrency changes, run `go test -race ./...` with a supported compiler/CGO toolchain. Report missing tooling instead of claiming a pass. Use `gofmt` on changed Go files and `golangci-lint run ./...` when available and relevant.

Repository integration tests are excluded from the ordinary suite by the `integration` build tag. Set `COLLECTION_TEST_POSTGRES_DSN` in the current shell to a disposable PostgreSQL test database, then run:

```powershell
go test -tags=integration ./internal/collection/repository
```

Without the DSN the test skips; a skip is not database validation. Existing tests cover collection handler pagination, service cache variants/invalidation and fallback, response serialization, shared error parsing, and shortlink validation/error paths. Extend relevant coverage when behavior changes.

The Makefile provides `migrate`, `test`, `lint`, and `mock` targets. Its test recipe relies on Make, `grep`, and `rm`, so prefer direct Go commands in native PowerShell. The mock target requires `mockery`; regenerate mocks only when needed.

For documentation-only changes, verify links, paths, commands, routes, and configuration names against source and inspect the diff. No new tests, application test run, live service calls, or migrations are required solely for prose changes.
