# DIEALOG Backend

Go backend for DIEALOG's digital legacy management service. Users register the accounts and digital assets they want to pass on, designate stakeholders and groups, configure post-mortem actions, and write last-word messages.

## Stack

- Go 1.27
- [Fiber v3](https://gofiber.io/) HTTP framework
- [Ent](https://entgo.io/) ORM with PostgreSQL
- [Viper](https://github.com/spf13/viper) configuration
- JWT (`golang-jwt/jwt/v4`), bcrypt
- testify + mockery
- Swagger via swaggo

Account passwords are not stored in PostgreSQL. They are delegated to the DIEALOG crypto service configured by `INTERNAL_CRYPTO_URL`.

## Requirements

- Go 1.27+
- PostgreSQL 17, or Docker
- CGO/gcc if you run the SQLite-backed repository tests

## Setup

```bash
cp .env.example .env
docker compose up -d postgres
```

Set `INTERNAL_SECRET_KEY` and `INTERNAL_CRYPTO_URL` in `.env` to match your crypto service. The example uses `http://crypto:8080`, which is the container-network address used by the deployment stack; for a locally running crypto service use its reachable host/port instead.

Then start the API:

```bash
make run
```

Notes:

- `config.NewEnv()` loads `.env` when present and also binds every key as an environment variable. Environment variables take precedence, so the same binary works locally and in containers. `.env` itself is optional.
- PostgreSQL is exposed on host port **5433** (container port 5432), matching `.env.example`.
- Ent runs `Schema.Create` on startup, so schema changes are applied automatically.
- The API listens on `SERVER_ADDRESS` (for example `:8080`). `PORT` in `.env.example` is not read by the app.

### Environment variables

| Variable | Description |
| --- | --- |
| `APP_ENV` | Environment name (`development` enables extra logging) |
| `SERVER_ADDRESS` | Fiber listen address, e.g. `:8080` |
| `CONTEXT_TIMEOUT` | Use-case timeout in seconds |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | PostgreSQL connection |
| `ACCESS_TOKEN_EXPIRY_HOUR` | Access-token lifetime in hours |
| `REFRESH_TOKEN_EXPIRY_HOUR` | Refresh-token lifetime in hours |
| `ACCESS_TOKEN_SECRET`, `REFRESH_TOKEN_SECRET` | JWT signing secrets |
| `INTERNAL_SECRET_KEY` | Secret sent to the crypto service as `X-Internal-Secret-Key` |
| `INTERNAL_CRYPTO_URL` | Base URL of the DIEALOG crypto service |

## API

All routes are under `/api`. Protected routes require:

```http
Authorization: Bearer <accessToken>
```

| Resource | Endpoint | Methods |
| --- | --- | --- |
| Auth | `/api/signup`, `/api/login`, `/api/refresh` | `POST` |
| Profile | `/api/profile/protected` | `GET` |
| Profile | `/api/profile/protected/image` | `PUT` (multipart field `profile_image`) |
| Profile | `/api/profile/protected/config` | `PUT` |
| Legacy | `/api/legacy/protected` | `POST`, `GET` |
| Legacy | `/api/legacy/protected/{id}` | `GET`, `PUT`, `DELETE` |
| Legacy account | `/api/legacy/protected/{id}/account` | `POST` |
| Legacy account | `/api/legacy/protected/{id}/account/{accountId}` | `GET`, `DELETE` |
| Last word | `/api/last-word/protected` | `POST`, `GET` |
| Last word | `/api/last-word/protected/{id}` | `PUT`, `DELETE` |
| Stakeholder | `/api/stakeholder/protected` | `POST`, `GET` |
| Stakeholder | `/api/stakeholder/protected/{id}` | `GET`, `PUT`, `DELETE` |
| Stakeholder group | `/api/stakeholder/protected/group` | `POST`, `GET` |
| Stakeholder group | `/api/stakeholder/protected/group/{id}` | `GET`, `PUT`, `DELETE` |
| Transition policy | `/api/transition/protected/policy` | `GET`, `PATCH` |
| Transition confirmer | `/api/transition/protected/confirmers` | `POST`, `GET` |
| Transition confirmer | `/api/transition/protected/confirmers/{id}` | `DELETE` |
| Transition confirmer | `/api/transition/protected/confirmers/{id}/resend` | `POST` |
| Alive check | `/api/transition/protected/alive-checks` | `POST`, `GET` |
| Alive check | `/api/transition/protected/alive-checks/{id}/respond` | `POST` |
| Transition case | `/api/transition/protected/cases` | `GET` |
| Transition case | `/api/transition/protected/cases/{id}` | `GET` |
| Transition case | `/api/transition/protected/cases/{id}/cancel` | `POST` |
| Transition case | `/api/transition/protected/cases/{id}/confirmations` | `GET` |
| Transition case | `/api/transition/protected/cases/{id}/tasks` | `GET` |
| Account handling task | `/api/transition/protected/tasks/{id}/request` | `POST` |
| Stakeholder task result | `/api/transition/stakeholder-tasks/result` | `POST` (JSON body `token`, `success`) |
| Confirmer flow | `/api/transition/confirmer/cases` | `POST`, `GET` |
| Confirmer flow | `/api/transition/confirmer/cases/{id}` | `GET` |
| Confirmer flow | `/api/transition/confirmer/cases/{id}/confirmations` | `POST` |
| Invitation | `/api/transition/invitations/accept` | `POST` (JSON body `token`) |
| Invitation | `/api/transition/invitations/decline` | `POST` (JSON body `token`) |

Confirmer routes require the token returned by the invitation acceptance flow:

```http
X-Confirmer-Token: <confirmerToken>
```

Invitation, confirmer auth, and stakeholder request tokens are returned only in the response that issues them because no email/SMS delivery service is configured. Only their SHA-256 hashes are stored. Stakeholder request tokens are bearer credentials; there is no stakeholder login yet, so identity is assumed from token possession. `Account.Action` is an instruction for the account's designated stakeholders, not an operation the backend performs: after the owner sends a request with `POST /api/transition/protected/tasks/{id}/request`, a stakeholder reports success or failure with `POST /api/transition/stakeholder-tasks/result`. Even the `remain` action stays pending until a stakeholder reports success. Any one assigned stakeholder result settles the task; the other pending requests are cancelled. Each request additively refreshes assignees from the account's current direct stakeholders and stakeholder-group members, retains existing assignees, and rotates tokens for non-confirmed assignees. There is no background scheduler or worker: alive-check status is advanced when checks are read, scheduled, or answered, and account handling requests are issued manually. Current implementation defaults are a 7-day invitation lifetime, a 30-day confirmer auth-token lifetime, a 7-day stakeholder request lifetime, and a 7-day alive-check response window; these are placeholders until product policy is finalized.

Interactive documentation:

- `/swagger` or `/docs` — Swagger UI
- `docs/swagger.json`, `docs/swagger.yaml` — generated specification

Uploaded profile images are served from `./public` at `/public/...` (for example `/public/profiles/<file>`).

## Development

```bash
make help            # list targets
make run             # go run ./cmd/main.go
go build ./...       # compile all packages
make test            # go test ./...
make ent-generate    # regenerate committed Ent code after editing ent/schema/*.go
make mock            # regenerate domain/mocks from domain interfaces
make swag            # regenerate docs/ from handler annotations
```

Code generation:

- Edit `ent/schema/*.go`, then run `make ent-generate`. Files under `ent/` are generated and committed; do not hand-edit them.
- Run `make mock` after changing interfaces in `domain/`.
- Swagger annotations live on handlers in `api/handler/`. Run `make swag` after changing them.

`go test ./...` runs the full suite, including SQLite-backed repository tests (CGO/gcc required).

## Project layout

```
api/         HTTP handlers, routes, JWT middleware
usecase/     business logic and validation
domain/      interfaces, DTOs, response mappers, errors
infra/       Ent repositories, database setup, crypto-service client
ent/         Ent schemas and generated ORM code
cmd/main.go  composition root
docs/        generated Swagger specification
```

Dependencies point inward: `api -> usecase -> domain`, with `infra` implementing `domain` interfaces. `cmd/main.go` wires concrete implementations manually.

## Docker

`docker compose up -d postgres` starts only the database, which is the recommended local setup.

`docker compose up -d --build` runs the full stack. Compose reads `.env` for interpolation and injects the configuration into the `api` container, overriding `DB_HOST` to `postgres` and `DB_PORT` to `5432`. The container does not need a `.env` file because the app also reads environment variables.

## Deployment

Pushes to `main` trigger a self-hosted deployment through the umbrella repository's submodules. CI builds `cmd/main.go` but does not run tests.
