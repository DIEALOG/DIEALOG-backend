# Go Backend Clean Architecture

> PR 은 언제나 환영합니다

## Framework / Library
- **Fiber**
- **Ent**
- **testify**
- **mockery**
- **viper**
- **PostgreSQL**
- **bcrypt**
- Check more packages in `go.mod`.

## How to Run

- clone 

```bash
cd your-workspace

git clone https://github.com/janghanul090801/go-backend-clean-architecture-fiber.git

cd go-backend-clean-architecture-fiber
```
- run
```bash
make run 
```

### Run Test

```bash
make test
```


### The Complete Project Folder Structure

```
.
├── api/
│   ├── handler/
│   │   ├── auth_handler.go
│   │   ├── profile_handler.go
│   │   ├── profile_handler_test.go
│   │   └── task_handler.go
│   ├── middleware/
│   │   └── jwt_auth_middleware.go
│   └── route/
│       ├── login_route.go
│       ├── profile_route.go
│       ├── refresh_token_route.go
│       ├── signup_route.go
│       └── task_route.go
├── cmd/
│   └── main.go
├── config/
│   └── env.go
├── domain/
│   ├── mocks/
│   │   ├── AuthUseCase.go
│   │   ├── ProfileUseCase.go
│   │   ├── TaskRepository.go
│   │   ├── TaskUseCase.go
│   │   └── UserRepository.go
│   ├── auth.go
│   ├── domain.go
│   ├── error_response.go
│   ├── error.go
│   ├── jwt_custom.go
│   ├── pagination.go
│   ├── profile.go
│   ├── success_response.go
│   ├── task.go
│   └── user.go
├── ent/
│   ├── schema/
│   │   ├── task.go
│   │   └── user.go
│   └── ... (generated files)
├── infra/
│   ├── database/
│   │   └── database.go
│   └── repository/
│       ├── mapper.go
│       ├── task_repository.go
│       ├── user_repository_test.go
│       └── user_repository.go
├── internal/
│   ├── fakeutil/
│   │   └── fakeutil.go
│   └── token/
│       └── token.go
├── usecase/
│   ├── auth_usecase.go
│   ├── profile_usecase.go
│   ├── task_usecase_test.go
│   └── task_usecase.go
├── .env.example
├── .gitignore
├── docker-compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── help.ps1
├── LICENSE
├── Makefile
└── README.md
```

### API Request & Response

- signup

  - request

  ```
  curl --location --request POST 'http://localhost:8080/api/signup' \
  --data-urlencode 'email=test@gmail.com' \
  --data-urlencode 'password=test' \
  --data-urlencode 'name=Test Name'
  ```

  - response

  ```json
  {
    "accessToken": "access_token",
    "refreshToken": "refresh_token"
  }
  ```

- login

  - request

  ```
  curl --location --request POST 'http://localhost:8080/api/login' \
  --data-urlencode 'email=test@gmail.com' \
  --data-urlencode 'password=test'
  ```

  - response

  ```json
  {
    "accessToken": "access_token",
    "refreshToken": "refresh_token"
  }
  ```

- profile

  - request

  ```
  curl --location --request GET 'http://localhost:8080/api/profile/protected' \
  --header 'Authorization: Bearer access_token'
  ```

  - response

  ```json
  {
    "name": "Test Name",
    "email": "test@gmail.com"
  }
  ```

- task create

  - request

  ```
  curl --location --request POST 'http://localhost:8080/api/task/protected' \
  --header 'Authorization: Bearer access_token' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'title=Test Task'
  ```

  - response

  ```json
  {
    "message": "Task created successfully"
  }
  ```

- task fetch

  - request

  ```
  curl --location --request GET 'http://localhost:8080/api/task/protected?page=1&limit=10' \
  --header 'Authorization: Bearer access_token'
  ```

  - response

  ```json
  {
    "items": [
      {
       "id": "626afac0-5e0d-493a-a97b-1827b485f521",
        "title": "Test Task",
        "user_id": "b10e63ff-6ab2-4714-8f90-c33a99e7b476",
        "created_at": "2026-02-11T16:36:01.839759Z"
      },
      {
        "id": "f46834d6-649c-4a28-b779-bcd74aad0a06",
        "title": "Test Task",
        "user_id": "b10e63ff-6ab2-4714-8f90-c33a99e7b476",
        "created_at": "2026-02-11T16:34:46.026427Z"
      }
    ],
    "total": 2,
    "limit": 5,
    "offset": 0
  }
  ```

- refresh token

  - request

  ```
  curl --location --request POST 'http://localhost:8080/api/refresh' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'refreshToken=refresh_token'
  ```

  - response

  ```json
  {
    "accessToken": "access_token",
    "refreshToken": "refresh_token"
  }
  ```

