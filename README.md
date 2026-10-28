# DIEALOG 백엔드

디지털 유산 관리 서비스 DIEALOG의 Go 기반 백엔드입니다. 사용자는 자신이 소유한 계정과 디지털 자산을 지정하고, 이해관계자 및 그룹을 구성하며, 사후 조치를 설정하고 유언 메시지를 작성할 수 있습니다.

## 기술 스택

- **Go 1.27**
- [Fiber v3](https://gofiber.io/) HTTP 프레임워크
- [Ent](https://entgo.io/) ORM와 PostgreSQL
- [Viper](https://github.com/spf13/viper) 설정
- JWT (`golang-jwt/jwt/v4`), bcrypt
- testify + mockery
- Swagger via swaggo

계정 비밀번호는 PostgreSQL에 저장되지 않습니다. DIEALOG crypto 서비스에 위임되며, `.env`에서 `INTERNAL_CRYPTO_URL`로 구성됩니다.

## 요구사항

- Go 1.27+
- PostgreSQL 17, 또는 Docker
- SQLite 기반 리포지토리 테스트를 실행하려면 CGO/gcc

## 설정

```bash
cp .env.example .env
docker compose up -d postgres
```

`INTERNAL_SECRET_KEY`와 `INTERNAL_CRYPTO_URL`을 `.env`에 설정을 하여 crypto 서비스와 일치시킵니다. 예시에서는 `http://crypto:8080`을 사용하며, 이는 배포 스택에서 컨테네트워크 주소입니다. 로컬에서 실행하는 crypto 서비스인 경우 reachable한 호스트/포트를 사용하세요.

그런 다음 API를 시작합니다:

```bash
make run
```

### 환경 변수

| 변수 | 설명 |
|------|------|
| `APP_ENV` | 환경 이름(`development`는 추가 로깅을 활성화) |
| `SERVER_ADDRESS` | Fiber 대기 주소, 예: `:8080` |
| `CONTEXT_TIMEOUT` | usecase 타임아웃(초) |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | PostgreSQL 연결 |
| `ACCESS_TOKEN_EXPIRY_HOUR` | 액세스 토큰 수명(시간) |
| `REFRESH_TOKEN_EXPIRY_HOUR` | 리프레시 토큰 수명(시간) |
| `ACCESS_TOKEN_SECRET`, `REFRESH_TOKEN_SECRET` | JWT 서명 비밀키 |
| `INTERNAL_SECRET_KEY` | crypto 서비스에 `X-Internal-Secret-Key`로 전송되는 비밀키 |
| `INTERNAL_CRYPTO_URL` | DIEALOG crypto 서비스 기본 URL |

## API

모든 라우트는 `/api` 아래에 있으며, 보호된 라우트는 인증이 필요합니다.

```
Authorization: Bearer <accessToken>
```

| 리소스 | 엔드포인트 | 메서드 |
|--------|------------|--------|
| Auth | `/api/signup`, `/api/login`, `/api/refresh` | `POST` |
| Profile | `/api/profile/protected` | `GET` |
| Profile | `/api/profile/protected/image` | `PUT` (multipart 필드 `profile_image`) |
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

확인자(confirmers) 경로는 초대 수락 흐름으로 반환된 토큰이 필요합니다:

```
X-Confirmer-Token: <confirmerToken>
```


인터랙티브 문서:

- `/swagger` 또는 `/docs` — Swagger UI
- `docs/swagger.json`, `docs/swagger.yaml` — 생성된 스펙

업로드된 프로필 이미지는 `./public`에서 `/public/...`으로 제공됩니다(예: `/public/profiles/<파일>`).

## 개발

```bash
make help            # 대상 목록 표시
make run             # go run ./cmd/main.go
go build ./...       # 모든 패키지 컴파일
make test            # go test ./...
make ent-generate    # ent/schema/*.go 편집 후 Ent 코드 재생성
make mock            # domain 인터페이스 변경 후 domain/mocks 재생성
make swag            # handler 어노테이션 변경 후 docs/ 재생성
```

## 프로젝트 구조

```
api/         HTTP 핸들러, 라우팅, JWT 미들웨어
usecase/     비즈니스 로직 및 검증
domain/      인터페이스, DTO, 응답 매퍼, 에러
infra/       Ent 리포지토리, 데이터베이스 설정, crypto 서비스 클라이언트
ent/         Ent 스키마 및 생성된 ORM 코드
cmd/main.go  구성 루트
docs/        생성된 Swagger 스펙
```

## Docker

- `docker compose up -d postgres`는 데이터베이스만 시작하며, 로컬 설정의 권장 방식입니다.
- `docker compose up -d --build`는 전체 스택을 실행합니다. Compose는 `.env`를 읽어 보간하고, `api` 컨테이너에 구성 정보를 주입하여 `DB_HOST`를 `postgres`, `DB_PORT`를 `5432`로 override합니다. 앱은 환경 변수도 직접 읽습니다.
