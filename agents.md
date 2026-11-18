# Hobby Collection Backend: Agent Guidance

## Start here

This Go HTTP API uses Gin, GORM, PostgreSQL, Redis, and Cloudinary. Collection and shortlink features use handler, service, repository, and entity packages; admin token issuance has a smaller handler-based module.

Read this file and the [workflow guide](.agents/workflow.md) before starting work. Read the relevant supporting guides before changing a subsystem; do not assume linked files are automatically loaded.

| Task | Read |
| --- | --- |
| Locate code or change domain logic and persistence | [Architecture and conventions](.agents/architecture.md) |
| Change routes, authentication, binding, or responses | [HTTP API guidance](.agents/api.md) |
| Configure, run, migrate, test, or change caching and integrations | [Development and validation](.agents/development.md) |
| Delegate exploration, investigate, or review changes | [Agent workflow](.agents/workflow.md) |

## Essential conventions

- Pass `context.Context` first in service/repository I/O operations, starting with `c.Request.Context()` in handlers. Pure helpers and in-memory cache invalidation do not need a synthetic context.
- Keep HTTP binding in handlers, business rules in services, and database queries in repositories. Follow existing interfaces and module wiring.
- Use request/response DTOs with explicit JSON/form tags. Preserve existing response shapes, including shortlink exceptions.
- Return request-processing errors rather than panicking. Use shared error types and response helpers where the endpoint does.
- Bind SQL values with GORM `?` placeholders; choose dynamic ordering from fixed supported choices.
- Review cache invalidation when changing collection writes, catalogs, filters, or pagination.
- Keep credentials out of source, documentation, and logs. Use `config.example.yml` to document configuration.
- Preserve unrelated working-tree changes. Validate in proportion to the change and report checks that could not run.

## Documentation maintenance

Update the relevant guide when behavior, configuration, commands, or architecture changes. Verify claims against source and distinguish current behavior from instructions for new changes. Keep this entry point short. Do not normalize public contracts as incidental cleanup.
