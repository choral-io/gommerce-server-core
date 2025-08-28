# Repository Guidelines

## Project Structure & Module Organization
- Module: `github.com/choral-io/gommerce-server-core` (library-first).
- Packages: `config/` (YAML loaders), `server/` (HTTP/gRPC wiring, health), `secure/` (JWT/Redis tokens, identity), `logging/` (zap/slog bridges, gRPC middleware, trace correlation), `otel/` (tracing/metrics), `data/` (Bun, Redis, Snowflake), `events/` (NATS), `dlock/` (locks), `validator/` (validation and interceptors).
- Root files: `go.mod`, `.env.example`, `.editorconfig`, `README.md`.

## Build, Test, and Development Commands
- Go: 1.25+.
- Build: `go build ./...`.
- Test: `go test ./... -v -cover`. Focused example: `go test -run Secure -v ./secure`.
- Vet/format: `go vet ./...`; `go fmt ./...`; `go mod tidy` before PRs.

## Coding Style & Naming Conventions
- Go style: `gofmt`. Package names are short lowercase nouns; exported identifiers use `CamelCase`.
- Indentation: default 4 spaces; YAML/JSON use 2 (see `.editorconfig`).
- Final newline: ensure files end with a newline (per `.editorconfig`).
- Errors: wrap with `%w`; define sentinels as `var ErrX = errors.New(...)`.
- APIs: accept `ctx context.Context` first; keep parameters concise.

## Testing Guidelines
- Framework: native `testing`. Files `*_test.go`; functions `TestXxx(t *testing.T)`.
- Prefer deterministic, table-driven tests and subtests (`t.Run`).
- Coverage: run `go test -cover ./...`; prioritize `secure/`, `server/`, and `config/`.

## Commit & Pull Request Guidelines
- Conventional commits: `feat:`, `fix:`, `refactor:`, `chore:`.
- Branch names: short kebab-case (e.g., `feat/http-mux`, `fix/token-ttl`).
- PRs: clear description, motivation, linked issues. Include tests and note config/API changes. Ensure `go build ./...` and `go test ./...` pass.

## Security & Configuration Tips
- Never commit secrets. Use `.env` locally; start from `.env.example` and `config/example.yaml`.
- Manage JWT keys via files or inline values; avoid real keys in VCS.
- Make endpoints (DB, Redis, NATS, OTEL) configurable for local and CI/CD environments.

## Architecture Overview
- Library-first: integrate packages into your `main` via `fx` and `config.ExtractSections`.
- Observability: enable tracing/metrics via `otel/` and logging presets; defaults are safe for local dev.
