# DIEALOG Backend — 사후 전환 도메인 구현 계획

## 1. 목적

현재 작성된 domain model을 기반으로 사용자 사후 전환 기능을 구현한다.

구현 전 기존 프로젝트의 디렉터리 구조, domain interface, Ent schema, repository, usecase, handler, route, middleware 및 response 규칙을 확인한다. 기존 코드 스타일과 아키텍처를 유지하며 구현한다.

### 핵심 도메인 규칙

1. 사후 전환 정책은 **계정별이 아닌 사용자별**이다.
2. 사용자의 사후 전환이 확정되면 해당 사용자가 소유한 모든 `Legacy`와 그 하위 `Account`가 처리 대상이 된다.
3. `User` → `Legacy` → `Account`의 기존 소유 관계를 유지한다.
4. 사후 전환 확인자와 계정 처리 담당자는 서로 다른 역할이다.
5. 사후 전환 정책과 확인자 관리는 `UserConfig`에서 분리한다.
6. `Account.Action` 및 `Account.ActionMessage`는 계정별 처리 방침으로 유지한다.
7. 각 계정의 실제 처리 작업은 독립적으로 진행한다. 한 계정의 실패가 다른 계정의 작업을 중단시키면 안 된다.
8. 생존 확인 기한 초과만으로 사망을 확정하거나 계정 처리 작업을 실행하면 안 된다.
9. 모든 식별자는 프로젝트에서 사용하는 UUID 기반 `domain.ID` 규칙을 따른다.

---

## 2. 구현 전 조사

코드를 수정하기 전에 다음 사항을 조사한다.

- [ ] 기존 domain model 및 UseCase 인터페이스
- [ ] `UserConfig`, `Stakeholder`, `StakeholderGroup`, `Legacy`, `Account`의 Ent schema
- [ ] 기존 Repository 구현과 domain mapper
- [ ] 기존 Handler, Route, JWT 인증 미들웨어
- [ ] 기존 API 응답 구조 및 오류 처리 방식
- [ ] 기존 `UserConfig.verifier`의 타입과 사용처
- [ ] `UserConfig.conversion_criteria`의 실제 의미와 사용처
- [ ] `UserConfig.survival_check_interval_days`의 사용처
- [ ] 기존 `AccountActionField` 및 `Account.ActionMessage` 처리 방식
- [ ] 프로젝트의 테스트, Swagger, 환경 설정 및 의존성 주입 방식

조사 결과를 바탕으로 기존 domain model과 중복되는 구조체나 인터페이스를 만들지 않는다. 필요한 경우 기존 모델을 확장하고, 새 도메인의 책임만 추가한다.

---

## 3. 도메인 모델 구현

### 3.1. UserTransitionPolicy

사용자 한 명의 사후 전환 정책을 관리한다.

필수 책임:

- 사용자별 정책 조회 및 생성
- 정책 변경
- 정책 값 검증
- 사용자당 정책 하나만 존재하도록 보장

주요 필드:

- `ID`
- `UserID`
- `MinimumConfirmers`
- `WaitingPeriodDays`
- `AliveCheckIntervalDays`
- `CreatedAt`
- `UpdatedAt`

초기 기본값:

- 최소 동의 확인자 수: 2명
- 대기 기간: 7일
- 생존 확인 주기: 90일

필드명과 타입은 이미 작성된 domain model을 우선 확인한다.

검증 규칙:

- 최소 확인자 수는 1 이상이어야 한다.
- 대기 기간은 음수일 수 없다.
- 생존 확인 주기는 양수여야 한다.
- 사용자가 다른 사용자의 정책을 조회하거나 수정할 수 없어야 한다.

`UserConfig`의 `conversion_criteria`, `survival_check_interval_days`는 해당 의미를 새 정책 필드에 반영한다. `conversion_criteria`의 기존 의미가 불명확하면 임의로 해석하지 말고 코드 사용처를 확인한 뒤 매핑한다.

### 3.2. UserTransitionConfirmer

사용자별 사후 전환 확인자 및 초대 상태를 관리한다.

확인자는 기존 `Stakeholder`와 연결할 수 있지만, 확인자로 지정되었다는 사실과 초대를 수락했다는 사실은 구분한다.

필요한 기능:

- 확인자 초대
- 확인자 목록 조회
- 초대 재발송
- 초대 수락 및 거절
- 확인자 철회
- 초대 만료 처리
- 중복 확인자 등록 방지

상태 예시:

- `pending`
- `accepted`
- `declined`
- `revoked`
- `expired`

구현 규칙:

- 초대 수락은 확인자 역할을 수락했다는 의미이며, 사후 전환 동의를 의미하지 않는다.
- 초대 수락 및 거절은 초대 대상의 신원을 검증한 뒤 처리한다.
- `StakeholderID`가 존재하면 해당 이해관계자가 초대 대상 사용자의 소유 범위에 속하는지 검증한다.
- 초대 대상이 기존 `Stakeholder`로 등록되지 않은 경우를 지원할지는 기존 domain model과 제품 요구사항을 확인한다.
- 동일 사용자의 확인자 중복 등록을 방지한다.
- 재초대 시 이전 초대 토큰을 무효화한다.
- 초대 토큰을 사용한다면 원본 토큰 대신 검증 가능한 해시를 저장하고 만료 시간을 관리한다.
- 초대 토큰은 추측하기 어려운 난수로 생성하고, 만료·재사용·중복 수락을 방지한다.
- 확인자 철회 후에도 과거 전환 확인 기록은 보존한다.

기존 `UserConfig.verifier`는 확인자 모델로 이전하는 것이 목표다. 기존 필드의 실제 타입과 참조 관계를 확인하고 새로운 모델에 적합하게 연결한다.

### 3.3. UserAliveCheck

사용자별 생존 확인 요청 및 응답 이력을 관리한다.

주요 필드:

- `ID`
- `UserID`
- `Sequence`
- `ScheduledAt`
- `DueAt`
- `RespondedAt`
- `Status`
- `CreatedAt`

상태 예시:

- `scheduled`
- `pending`
- `responded`
- `overdue`
- `cancelled`

필수 기능:

- 다음 생존 확인 일정 계산
- 생존 확인 요청 생성
- 확인 요청 발송
- 응답 처리
- 기한 초과 처리
- 재알림
- 사용자별 생존 확인 이력 조회

구현 규칙:

- 생존 확인 주기는 `UserTransitionPolicy.AliveCheckIntervalDays`를 기준으로 한다.
- 주기 변경은 이후 일정부터 적용하고 이미 발송된 요청의 기한을 임의로 변경하지 않는다.
- 기한 초과는 사망 확정이 아니다.
- 생존 확인 요청 및 응답 이력을 보존한다.
- 스케줄러가 재실행되어도 동일한 일정의 확인 요청이 중복 생성되지 않도록 한다.

### 3.4. UserTransitionCase

사용자 한 명의 사후 전환 절차 한 건을 나타낸다.

주요 필드:

- `ID`
- `UserID`
- `Status`
- `MinimumConfirmersSnapshot`
- `WaitingPeriodDaysSnapshot`
- `RequestedAt`
- `WaitingStartedAt`
- `EligibleAt`
- `ConfirmedAt`
- `CancelledAt`
- `CreatedAt`
- `UpdatedAt`

상태 예시:

- `pending_confirmation`
- `waiting_period`
- `confirmed`
- `cancelled`

필수 기능:

- 사후 전환 절차 생성
- 절차 상태 조회
- 확인자 동의 수 집계
- 동의 기준 충족 여부 판단
- 대기 기간 관리
- 사후 전환 확정
- 절차 취소
- 과거 절차 이력 조회

구현 규칙:

- 동시에 진행할 수 있는 활성 전환 절차는 사용자당 최대 하나다.
- 절차 시작 시 정책의 최소 확인자 수와 대기 기간을 스냅샷으로 저장한다.
- 현재 등록된 확인자 수가 아니라 유효한 동의 응답 수를 기준으로 판단한다.
- 동의 기준 충족 후 대기 기간이 끝나기 전에는 확정하지 않는다.
- 절차가 취소되면 관련 이력을 삭제하지 않는다.
- 상태 전이는 명시적으로 검증한다.
- 동시 요청으로 중복 절차가 생성되지 않도록 한다.
- 단순한 확인자 동의나 생존 확인 기한 초과만으로 무조건 사후 전환을 확정하지 않는다.

사후 전환의 최종 확정 조건과 취소 가능 시점은 명시적인 정책으로 구현한다. 정의되지 않은 조건을 임의로 추가하지 않는다.

### 3.5. UserTransitionConfirmation

특정 전환 절차에서 확인자별로 제출한 응답을 관리한다.

주요 필드:

- `ID`
- `TransitionCaseID`
- `TransitionConfirmerID`
- `Decision`
- `Comment`
- `RespondedAt`
- `CreatedAt`

응답 예시:

- `agree`
- `disagree`
- `abstain`

필수 기능:

- 확인자 응답 등록
- 응답 조회
- 응답 변경 규칙 적용
- 절차별 동의 인원 집계
- 확인자 권한 검증

구현 규칙:

- 해당 사용자의 확인자로 지정된 사람만 응답할 수 있다.
- 확인자 초대를 수락하지 않은 사용자의 응답을 허용하지 않는다.
- 같은 확인자가 동일 절차에서 중복 응답을 생성하지 않도록 한다.
- 응답 변경 및 철회는 명시적인 정책에 따라 처리한다.
- 과거 응답 이력을 임의로 삭제하지 않는다.
- 확인자 철회가 기존 응답에 미치는 영향은 명시적인 규칙으로 구현한다.

---

## 4. 계정 처리 도메인

### 4.1. AccountHandlingTask

사용자 사후 전환이 확정되었을 때 각 `Account`에 생성하는 독립적인 작업이다.

주요 필드:

- `ID`
- `TransitionCaseID`
- `AccountID`
- `ActionSnapshot`
- `ActionMessageSnapshot`
- `Status`
- `StartedAt`
- `CompletedAt`
- `FailureReason`
- `RetryCount`
- `CreatedAt`
- `UpdatedAt`

작업 상태 예시:

- `pending`
- `in_progress`
- `completed`
- `failed`
- `skipped`

필수 기능:

- 전환 확정 후 계정별 작업 생성
- 작업 상태 조회
- 작업 실행
- 실패 사유 기록
- 재시도
- 완료 처리
- 계정별 작업 목록 조회

작업 생성 규칙:

1. 확정된 `UserTransitionCase`의 소유자를 확인한다.
2. 해당 사용자가 소유한 모든 `Legacy`를 조회한다.
3. 각 `Legacy`의 모든 `Account`를 조회한다.
4. 각 계정의 `Action` 및 `ActionMessage`를 스냅샷으로 저장한다.
5. 각 계정에 작업을 생성한다.

제약:

- `(TransitionCaseID, AccountID)`는 유일해야 한다.
- 다른 사용자의 계정을 작업 대상으로 지정할 수 없다.
- 계정 하나의 실패가 다른 계정 작업을 중단시키면 안 된다.
- 작업 재시도는 멱등성을 보장해야 한다.
- 작업 생성 도중 오류가 발생하더라도 확정된 절차의 미처리 계정이 누락되지 않도록 한다.
- 외부 계정 서비스 호출은 DB 트랜잭션과 분리한다.
- 작업 실행 전에 해당 전환 절차가 유효하게 확정되었는지 검증한다.

`Account.Action`은 처리 방침을 나타내며, `AccountHandlingTask.Status`는 실제 실행 상태를 나타낸다. 두 개념을 혼용하지 않는다.

### 4.2. AccountHandlingEvent

계정 처리 작업의 진행 이력을 기록한다.

필요한 기록:

- 작업 시작
- 작업 상태 변경
- 백업 시작 및 결과
- 삭제 또는 유지 처리 결과
- 실패 사유
- 재시도
- 완료

이벤트는 과거 기록을 덮어쓰기보다 추가하는 방식으로 관리한다.

### 4.3. 백업 후 삭제

기존 `backup-and-delete` 작업은 백업과 삭제가 서로 다른 단계라는 점을 고려한다.

초기 구현에서는 `AccountHandlingTask` 내부에서 단계별 실행 상태를 관리할 수 있다. 다만 다음 사항을 고려한다.

- 백업 성공 여부 검증
- 백업 실패 시 삭제 금지
- 백업 성공 후 삭제 실패 시 안전한 재시도
- 외부 서비스의 부분 실패 기록
- 재시도 시 중복 백업 및 중복 삭제 방지

기존 domain model에 작업 단계 모델이 없다면 먼저 현재 구조를 확인하고, 단계 분리가 필요할 때만 추가한다.

---

## 5. 기존 도메인과의 연결

### User

- 기존 사용자 엔티티를 유지한다.
- `UserTransitionPolicy`와 1:1 관계를 가진다.
- `UserTransitionConfirmer`, `UserAliveCheck`, `UserTransitionCase`와 1:N 관계를 가진다.
- 기존 `Legacy`, `Stakeholder`, `StakeholderGroup`, `UserConfig` 관계를 유지한다.

### Legacy 및 Account

- 기존 `User` → `Legacy` → `Account` 소유 관계를 유지한다.
- 계정별 처리 방침은 `Account.Action`과 `Account.ActionMessage`에 유지한다.
- 사후 전환 정책이나 생존 확인 주기를 `Account`에 중복 저장하지 않는다.

### Stakeholder 및 StakeholderGroup

- 기존 계정 처리 관련 이해관계자 관계를 유지한다.
- 사후 전환 확인자 역할은 `UserTransitionConfirmer`로 별도 관리한다.
- 확인자와 계정 처리 담당자를 같은 역할로 간주하지 않는다.
- `Stakeholder` 또는 `StakeholderGroup`과 연결할 때 소유권을 검증한다.

### UserConfig

유지할 설정:

- `lock_by_face_id`
- `certification_before_sensitive_information`

새 도메인으로 분리할 설정:

- `verifier`
- `conversion_criteria`
- `survival_check_interval_days`

기존 `conversion_criteria`의 의미는 실제 코드 사용처를 조사한 후 새 정책 필드에 매핑한다. 개발 단계이므로 DB migration 작업은 수행하지 않는다.

### LastWord

기존 사용자 소유의 독립된 도메인을 유지한다. 사후 전환 확정 시 편지 공개가 필요한 경우 별도 정책과 실행 규칙으로 구현하며, `LastWord`를 임의로 `Account` 또는 `Legacy`의 하위 엔티티로 변경하지 않는다.

---

## 6. Repository 및 mapper

새 도메인 모델마다 필요한 Repository 인터페이스와 구현을 추가한다.

대상:

- `UserTransitionPolicyRepository`
- `UserTransitionConfirmerRepository`
- `UserAliveCheckRepository`
- `UserTransitionCaseRepository`
- `UserTransitionConfirmationRepository`
- `AccountHandlingTaskRepository`
- `AccountHandlingEventRepository`

구현 원칙:

- 기존 domain interface와 Ent repository의 패턴을 따른다.
- Ent 엔티티와 domain model 사이의 mapper를 구현한다.
- 기존 mapper가 제공하는 nil 처리 및 관계 변환 규칙을 따른다.
- 여러 엔티티를 함께 생성하거나 상태를 변경할 때 필요한 트랜잭션 경계를 정의한다.
- 단순 CRUD를 넘어 동시성, 소유권, 상태 전이 제약을 고려한다.
- 기존 프로젝트에서 Repository가 여러 도메인 작업을 묶어 처리하는 규칙이 있다면 그 규칙을 우선한다.

---

## 7. UseCase

다음 책임을 기존 프로젝트의 UseCase 구조에 맞게 구현한다.

### Policy UseCase

- 정책 조회
- 정책 수정
- 정책 값 검증
- 기본 정책 생성

### Confirmer UseCase

- 확인자 추가
- 초대 발송 및 재발송
- 초대 수락 및 거절
- 확인자 철회
- 초대 만료
- 중복 초대 방지

### AliveCheck UseCase

- 생존 확인 일정 생성
- 요청 발송
- 응답 처리
- 기한 초과 및 재알림 처리

### TransitionCase UseCase

- 전환 절차 생성
- 활성 절차 중복 방지
- 동의 기준 판단
- 대기 기간 판단
- 확정 및 취소
- 절차 상태 조회

### TransitionConfirmation UseCase

- 확인자 권한 검증
- 응답 생성 및 변경
- 유효 동의 수 집계

### AccountHandling UseCase

- 확정된 전환 절차에서 모든 대상 계정 수집
- 계정별 작업 생성
- 작업 실행 및 결과 기록
- 실패 작업 재시도
- 작업별 상태 조회

UseCase는 Handler에 비즈니스 로직을 직접 작성하지 않고, domain interface를 통해 Repository와 외부 서비스를 호출한다.

---

## 8. Handler 및 API

기존 Fiber v3 라우트, 인증 미들웨어, 응답 DTO, 오류 처리 규칙을 먼저 확인하고 API를 추가한다.

권장 기능:

| 기능 | HTTP | 경로 예시 |
|---|---|---|
| 정책 조회 | GET | `/transition/policy` |
| 정책 변경 | PATCH | `/transition/policy` |
| 확인자 목록 | GET | `/transition/confirmers` |
| 확인자 초대 | POST | `/transition/confirmers` |
| 초대 재발송 | POST | `/transition/confirmers/{id}/resend` |
| 확인자 철회 | DELETE | `/transition/confirmers/{id}` |
| 초대 수락 | POST | `/transition/invitations/{token}/accept` |
| 초대 거절 | POST | `/transition/invitations/{token}/decline` |
| 생존 확인 응답 | POST | `/transition/alive-checks/{id}/respond` |
| 전환 절차 조회 | GET | `/transition/cases/{id}` |
| 전환 확인 응답 | POST | `/transition/cases/{id}/confirmations` |
| 계정별 작업 목록 | GET | `/transition/cases/{id}/tasks` |
| 작업 재시도 | POST | `/transition/tasks/{id}/retry` |

위 경로는 제안이며 기존 프로젝트의 라우팅 네이밍 규칙에 맞춰 조정한다.

필수 보안 규칙:

- 로그인 사용자는 본인의 정책과 전환 절차만 관리할 수 있다.
- 확인자는 본인이 권한을 가진 절차에만 응답할 수 있다.
- UUID 형식 검증과 리소스 소유권 검증을 수행한다.
- 초대 토큰 인증과 일반 JWT 인증을 명확하게 구분한다.
- 민감한 초대 토큰 및 계정 인증 정보를 로그에 기록하지 않는다.
- 외부 계정 처리 서비스의 오류를 프로젝트의 표준 오류 응답으로 변환한다.

모든 API에 기존 프로젝트의 Swagger 주석 규칙을 적용한다.

---

## 9. 테스트

### Unit Test

- 정책 필드 검증
- 기본 정책 생성
- 확인자 상태 전이
- 초대 만료 및 재발송
- 최소 동의 인원 계산
- 대기 기간 계산
- 전환 절차 상태 전이
- 생존 확인 기한 초과 처리
- 계정 처리 방침 스냅샷
- 작업 재시도 및 멱등성

### Integration Test

- 사용자별 정책 생성 및 조회
- 확인자 초대부터 수락까지의 흐름
- 확인자 동의에 따른 전환 절차 진행
- 대기 기간 종료 후 확정
- 사용자의 모든 Legacy 및 Account 작업 생성
- 일부 계정 실패 시 나머지 작업의 정상 진행
- 중복 요청 및 동시 확정 방지
- 다른 사용자의 리소스 접근 차단

### API Test

- 성공 및 실패 응답 형식
- 잘못된 UUID
- 인증 누락
- 권한 없는 확인자 응답
- 만료된 초대 토큰
- 중복 초대 및 중복 작업 요청
- 존재하지 않는 리소스
- 유효하지 않은 상태 전이

테스트는 기존 프로젝트의 테스트 프레임워크와 패턴을 사용한다. 외부 서비스 의존성은 테스트 가능한 인터페이스로 분리한다.

---

## 10. 구현 순서

1. **프로젝트 조사:** 현재 domain model, Ent schema, Repository, UseCase, Handler, Route를 파악한다.
2. **도메인 정합성 확인:** 새 모델과 기존 모델의 관계 및 상태 enum을 확정한다.
3. **Ent schema 및 mapper:** 필요한 스키마와 domain 변환을 구현한다.
4. **Repository:** 인터페이스와 Ent 기반 구현을 추가한다.
5. **UserTransitionPolicy:** 정책 조회·수정 및 기본값을 구현한다.
6. **UserTransitionConfirmer:** 확인자 등록, 초대, 수락, 거절, 철회를 구현한다.
7. **UserAliveCheck:** 일정 생성, 응답, 기한 초과 처리를 구현한다.
8. **UserTransitionCase 및 Confirmation:** 확인 절차, 동의 집계, 대기, 확정 및 취소를 구현한다.
9. **AccountHandlingTask 및 Event:** 확정 후 모든 계정의 작업 생성과 독립 실행을 구현한다.
10. **Handler 및 Route:** API, 인증, 오류 응답, Swagger를 구현한다.
11. **테스트 및 검증:** 단위·통합·API 테스트를 실행하고 실패를 수정한다.
12. **최종 보고:** 변경 파일, 구현 기능, 테스트 결과, 미구현 항목 및 추가 결정이 필요한 사항을 보고한다.

개발 단계이므로 데이터 migration은 구현 범위에서 제외한다.

---

## 11. Agent 작업 지침

- 구현 전 현재 코드를 조사하고 실제 파일과 인터페이스를 기준으로 작업한다.
- 기존 프로젝트의 아키텍처를 임의로 변경하지 않는다.
- 기존 domain model을 중복 생성하지 않는다.
- 필요한 파일만 수정하고, 불필요한 대규모 리팩터링을 하지 않는다.
- 스키마 생성 및 코드 생성이 필요한 경우 기존 프로젝트의 명령과 방식을 확인한 후 실행한다.
- 외부 API, 이메일·알림 발송, 계정 처리 서비스의 실제 구현 여부를 확인하고, 설정되지 않은 외부 기능을 성공한 것처럼 처리하지 않는다.
- 보안상 중요한 동작은 임시 하드코딩이나 우회 로직으로 구현하지 않는다.
- 요구사항이 명확하지 않은 정책은 임의로 확정하지 말고 구현 전 질문하거나 미결정 사항으로 보고한다.
- 작업 완료 후 실제 테스트 실행 결과를 보고한다. 실행하지 않은 테스트를 통과했다고 보고하지 않는다.

## 완료 기준

- 사용자별 사후 전환 정책을 관리할 수 있다.
- 확인자를 초대하고 초대 상태를 관리할 수 있다.
- 생존 확인 요청과 응답 이력을 관리할 수 있다.
- 확인자 응답과 대기 기간에 따라 전환 절차를 관리할 수 있다.
- 전환 확정 시 해당 사용자의 모든 계정에 작업이 생성된다.
- 각 계정의 작업은 독립적으로 실행되고 실패·재시도·완료 이력을 남긴다.
- 기존 인증, 권한 검증, API 응답, 오류 처리 및 테스트 규칙을 유지한다.
- 데이터 migration 없이 현재 개발 환경에서 구현 및 테스트할 수 있다.
