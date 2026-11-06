# AGENTS.md

`README.md` has the setup, environment-variable, and endpoint tables. This file records non-obvious constraints for coding agents.

## Project

DIEALOG backend: Go 1.27 REST API for digital legacy management. Fiber v3 + Ent (PostgreSQL) + Viper. Dependencies point inward: `api/{handler,route,middleware}` -> `usecase` -> `domain` (entities/DTOs + interfaces) -> `infra/{repository,client/crypto}`. `cmd/main.go` is the only composition root and all wiring is manual; adding a feature usually touches `domain/` interfaces, repository/usecase, handler/route, and `main.go`.

`implementation-plan.md` describes the post-transition design that is now implemented under `/api/transition`; keep its unresolved policy items explicit rather than silently inventing rules. Its account-executor/retry sections predate the stakeholder-confirmation correction; the implemented account flow is the stakeholder request/result flow below.

The transition flow is token-based. Owner routes stay under `/api/transition/protected/...`; invited confirmers use `/api/transition/confirmer/...` with `X-Confirmer-Token`; public invitation acceptance/decline uses `/api/transition/invitations/accept|decline` with the invitation token in the JSON body so it is not written to access logs.

- Invitation, confirmer auth, and stakeholder request tokens are returned only in the response that issues them because no email/SMS delivery service is configured. Only SHA-256 hashes are stored (`internal/securetoken`), and hashes must stay hidden from JSON.
- Account handling is a request to the account's designated stakeholders, not an action the backend executes. `Account.Action` and `ActionMessage` are snapshotted into the task and sent to direct account stakeholders plus members of the account's stakeholder groups. The task also snapshots the account detail (`LegacyID`, `Type`, `Name`, external account identifier as `account_identifier`, `DataLocation`) at creation; `AccountPassword` is never stored on the task and is resolved from the crypto service with the owner's user ID and the account ID when tasks are read. Even `remain` must be reported as successful by a stakeholder. Stakeholder request tokens are bearer credentials; there is no stakeholder login yet, so the backend cannot verify identity beyond token possession.
- The deceased owner's account stakeholders can issue or re-issue stakeholder request tokens with `POST /api/transition/protected/tasks/{id}/request` using their own JWT, because the owner cannot authenticate after death. A requester is authorized when their user ID is linked to one of the task's assignee snapshots or to a current account stakeholder (including group stakeholders) via `Stakeholder.stakeholder_id`; the owner remains authorized while alive. Each call additively refreshes assignees from the account's current stakeholders, retains existing assignees, and rotates outstanding tokens for non-confirmed assignees. Stakeholders report success or failure with `POST /api/transition/stakeholder-tasks/result`, sending the request token in the JSON body. Any one assignee result completes or fails the task; other pending requests are cancelled.
- `GET /api/transition/protected/cases/{id}/tasks` follows the same stakeholder scope: the owner sees every task, while a linked stakeholder sees only tasks for accounts they belong to (directly or through a stakeholder group) or where they are an assignee snapshot. Listing evaluates the case with the resolved owner ID, so a stakeholder can trigger confirmation/task creation without owner credentials.
- Transition confirmation is evaluated on demand (`Get`, list, submit, task list) and after a confirmer response. Tasks and their initial stakeholder assignee snapshots are created transactionally when a case moves from `waiting_period` to `confirmed`.
- Confirmers can read a case's decisions with `GET /api/transition/confirmer/cases/{id}/confirmations` using `X-Confirmer-Token`, because the owner cannot authenticate after death. The case is scoped to the confirmer's deceased user (`confirmer.UserID`), so another user's case returns 404. The owner JWT route (`/transition/protected/cases/{id}/confirmations`) is still available while the owner is alive.
- There is no background scheduler or worker yet. Alive-check status is advanced when checks are read, scheduled, or answered, and account handling requests are issued manually through the request endpoint.
- `UserTransitionPolicy` is the source of truth for confirmation/alive-check settings. `UserConfig.survival_check_interval_days` is only consulted when the initial policy is created; legacy `UserConfig` verifier/`conversion_criteria` fields are not migrated.

## Setup and run

- Run commands from the repo root. `config.NewEnv()` reads `.env` relative to the current directory, and profile uploads are written to `./public/profiles`.
- `cp .env.example .env`, `docker compose up -d postgres` (host port **5433**), then `make run` (`:8080`). Environment variables override `.env`; use `SERVER_ADDRESS`, not the unused `PORT` in `.env.example`. Full Compose sets the API's DB host/port to `postgres:5432`.
- The `Makefile` exports `APP_ENV=test` for every target, so `APP_ENV=development` from `.env` does not apply under `make` (the development log line is skipped).
- Legacy-account credentials go to the crypto service at `INTERNAL_CRYPTO_URL` with `X-Internal-Secret-Key`; the example `http://crypto:8080` only resolves inside Compose. User login passwords are bcrypt-hashed in `ent.User.password` in PostgreSQL.
- Startup runs `client.Schema.Create` (Ent, debug logging always on); there are no migration files to edit.
- The runtime Docker image runs as the non-root `appuser` (**uid 100 / gid 101**, not 1000) and pre-creates `/app/public/profiles` with that ownership; `/app` itself stays root-owned, so `public/` cannot be created at runtime by any other user. The Compose `api` service mounts the named volume `profile-uploads:/app/public`, so uploaded profile images persist across container rebuilds (but the named volume survives `docker compose up --build`; delete it if it was created root-owned by an older image).
- Do not override the container user with a different UID (for example `user: "1000:1000"` in an external Compose file). The app writes `./public/profiles` relative to the working directory, so a mismatch makes `os.MkdirAll`/`os.OpenFile` fail with "permission denied" and `PATCH`/profile-image uploads return 500. Either drop the override or set it to `100:101` to match the image.

## Verification

- Baseline: `go build ./...` and `go test ./...`. Focused tests: `go test ./usecase -run TestTransitionFlow` and `go test ./usecase -run TestFetchLegacyByUserID`.
- `go test ./...` needs CGO/gcc because repository tests use in-memory SQLite (`enttest` + `mattn/go-sqlite3`). Tests do not need PostgreSQL or the crypto service; keep new tests hermetic with mocks/`enttest`.
- `make integration-test` (`-tags integration -run TestIntegration`) exists, but there are currently no integration-tagged or `TestIntegration` tests, so it verifies nothing.
- CI runs only `go build ./cmd/main.go` on pushes and PRs to `main`; it does not run tests despite the job name.

## Codegen

- Ent: edit `ent/schema/*.go`, then `make ent-generate`. Everything under `ent/` is generated and committed; never hand-edit it.
- Mocks: change interfaces in `domain/`, then `make mock` (mockery v2.53.8 via `go run`). Don't hand-edit `domain/mocks`.
- Swagger: annotations live on handlers in `api/handler/` and are in Korean. `make swag` regenerates committed `docs/`. The target assumes a `swag` binary on `PATH`; if it is missing, run `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/main.go -o docs`.

## Conventions and gotchas

- This is Fiber **v3**, not v2: use `c.Bind().Body(...)`, `c.RequestCtx()` as the context, and `fiber.Ctx`. Don't paste Fiber v2 examples.
- `main.go` configures no `StructValidator`, so `validate:`/`binding:` tags on domain DTOs are inert. Handlers decode; validation lives in usecases (for example `Account.IsValid()` and the auth/stakeholder checks).
- Protected routes are `/api/<resource>/protected/...`; `JwtMiddleware` stores a `*domain.ID` in `c.Locals("id")`. `domain.ID` is a `uuid.UUID` alias, and interfaces pass `*domain.ID`. Parse path UUIDs with `domain.StringToID` and return 400 on failure.
- Route order matters: in `api/route/stakeholder_route.go`, keep `/protected/group` before `/protected/:id`; in `cmd/main.go`, keep `/public/*` before `app.All("*")`.
- Return `domain.NewBadRequestError` / `NewNotFoundError` / etc. and let `handler.RespondError` map `*domain.Error`; use Go 1.27's `errors.AsType[*domain.Error]` (pointer form). Non-domain errors become 500.
- Status conventions: CRUD creates return 201 (auth endpoints return 200); updates and deletes return 200, with deletes returning `domain.SuccessResponse` (not 204). Handled errors use `{"message": ...}`.
- Usecases wrap repository calls in `context.WithTimeout(c, timeout)` and take a `userID`. When adding repository methods, scope queries by owner edges (for example `legacy.HasOwnerWith(user.IDEQ(...))`) instead of trusting IDs.
- Auth handlers must `return RespondError(...)` immediately after a service error; falling through would create tokens with a nil user.
- Profile image uploads are limited to 5 MB and JPEG/PNG/GIF, stored with UUID filenames under `public/profiles`, and exposed at `/public/profiles/...`.
- Last-word receivers are `Stakeholder` rows, not users. `GET /api/last-word/protected/to-me/{id}` passes the authenticated user ID as the receiver and matches `Stakeholder.stakeholder_id` directly or through a receiver group, so a user only sees last words where one of their linked stakeholder contacts (or a group containing one) is a receiver.
