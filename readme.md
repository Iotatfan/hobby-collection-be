## Hobby Collection API (BE)

Go + Gin backend for the Hobby Collection site.

## Local run

```powershell
go run .\cmd\app\main.go
```

## Configuration

This service reads `config.yml` from the repo root (and also supports env var overrides using `_` for nesting, e.g. `POSTGRES_HOST`).

If `config.yml` does not exist, the app will create an empty one for you on startup—you still need to fill it in.

### `config.yml` template

## TODO

- [ ] Signed upload endpoint
- [ ] Move upload logic to FE
- [ ] Update edit & delete image logic