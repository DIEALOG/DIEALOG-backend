# AGENTS.md

## Project

DIEALOG backend: Go REST API for digital legacy management. Fiber v3 + Ent ORM (PostgreSQL) + Viper config. Layering is `api/{handler,route,middleware}` -> `usecase` -> `domain` (interfaces/DTOs) -> `infra/{repository,client/crypto}`. `cmd/main.go` is the only composition root; wiring is manual.

## Setup and run

- Go 1.27. `config.NewEnv()` loads `.env` when present and binds every config key as an environment variable; env vars win, and `.env` is optional.
- Postgres: `docker compose up -d postgres`; host port is **5433** (container 5432), matching `.env.example`. Then `make run` serves on `:8080`.
- Startup calls `client.Schema.Create(...)`, so Ent migrations run automatically. Ent debug logging is always enabled.
- Compose injects config into the `api` container via environment variables and overrides `DB_HOST=postgres`, `DB_PORT=5432`; no `.env` file is needed inside the image.

## Verification

- `go build ./...` and `go test ./...` both pass; use them as the baseline checks.
- Repository tests use in-memory SQLite via `enttest` and `mattn/go-sqlite3`, so CGO/gcc is required.
- CI (`.github/workflows/build.yml`) only runs `go build ./cmd/main.go`—it does not run tests despite the job name. Deployment is self-hosted, triggered by pushes to `main`, and happens through the umbrella repo's submodules.

## Codegen

- Ent: edit `ent/schema/*.go`, then `make ent-generate` (`go generate ./ent`). Everything under `ent/` is generated and committed; don't hand-edit it.
- Mocks: `make mock` regenerates `domain/mocks` from `domain` interfaces via mockery v2.53.8 (`go run`, no config file). Regenerate after changing interfaces in `domain/`.
- Swagger: `make swag` regenerates `docs/` from handler annotations (annotations are in Korean). Keep `docs/` in sync when routes or request types change.

## Conventions and gotchas

- This is Fiber **v3**, not v2: handlers use `c.Bind().Body(...)`, `c.RequestCtx()` as `context.Context`, and `fiber.Ctx`. Don't paste Fiber v2 examples.
- Protected routes are `/api/<resource>/protected/...`; `JwtMiddleware` stores a `*domain.ID` in `c.Locals("id")`, and handlers read it as `c.Locals("id").(*domain.ID)`.
- Return `domain.NewBadRequestError` / `NewNotFoundError` / etc. and let `handler.RespondError` map `*domain.Error` to status codes. Use Go 1.27's `errors.AsType[*domain.Error]` (pointer form)—the value form does not match service errors.
- Usecases wrap repository calls in `context.WithTimeout(c, timeout)` and take a `userID`; repositories enforce ownership through Ent owner edges (e.g. `legacy.HasOwnerWith(user.IDEQ(...))`).
- Account passwords are never stored in Postgres; they are sent to the sibling crypto service via `infra/client/crypto` with `X-Internal-Secret-Key`. Store expects HTTP 201; retrieve expects XML.
- Static files are served from `./public` at `/public/*` (profile images use `/public/profiles/...`). Keep the static route registered before `app.All("*")` in `cmd/main.go`.
- Auth handlers must return immediately after a service error via `RespondError`; falling through would call token creation with a nil user.

# Project Guidelines
- Follow the existing Clean Architecture structure.
- Use the existing domain interfaces and repository patterns.
- Keep HTTP handlers thin and delegate business logic to use cases.
- Use UUID for entity identifiers.
- Follow existing error response conventions.
- Add or update tests for changed behavior.
- Do not modify unrelated files.
- Run go test ./... after changes.