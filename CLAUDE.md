# CLAUDE.md

go-gin-lambda-starter — Go/Gin/GORM/Lambda 백엔드 스타터킷.

## 개발 환경 초기 설정

새 컴퓨터에서 처음 개발할 때 한 번 실행한다.

```bash
npm install
npm run setup    # goimports + golangci-lint 설치 + git 훅 활성화
```

## 빌드 & 실행

```bash
npm run local                     # 로컬 개발 서버 (localhost:8080)
npm run build                     # Lambda 바이너리 빌드 (linux/arm64)
npm run deploy:dev                # dev 환경 배포
npm run deploy:prod               # prod 환경 배포
```

## 코드 품질 (커밋 전)

```bash
npm run check                     # 포맷(goimports+gofmt) + 린트(go vet+golangci-lint)
npm run fmt                       # 포맷만
npm run lint                      # 린트만
npm test                          # 테스트
```

## 코딩 컨벤션

- Go 식별자: CamelCase (Go 표준)
- DB 컬럼: snake_case (GORM 자동 변환)
- DB 테이블명: snake_case + 단수형 (예: `item`, `member`, `api_log`)
- JSON 필드명: camelCase
- 주석·로그·커밋 메시지: 한국어
- 코드 계층: handler → service → dao (3계층). 자세한 구조는 [docs/architecture.md](docs/architecture.md) 참고.
- 파일 내 함수 순서: public 함수 먼저, 소문자(unexported) 함수는 `// --- Private 함수 ---` 구획 아래 파일 맨 끝에 배치
- Go visibility: 대문자 = exported (외부 패키지 접근 가능), 소문자 = unexported (패키지 내부 전용)

## 패키지 구조

- `com/` — 공통 인프라 (config, database, middleware, named_query, time_dao, response, dbtx) + 횡단 모델(`ApiLog`)
- `<feature>/` — 피처별 패키지 (handler/service/dao+sql + 도메인 모델). 예: `example/`
- `main.go` — 진입점 + 라우트 등록

## 모델 컨벤션

- 횡단 공통 모델만 `com/models.go` 에 (스타터킷 기본: `ApiLog`)
- 피처 도메인 모델은 해당 피처 패키지에 (예: `example/example_dao.go` 의 `Item`)
- 단일 PK: `id` BIGSERIAL auto-increment
- `MemberID`: 인증 미들웨어가 컨텍스트에 세팅한 sub (Firebase UID 등)
- nullable 필드: 포인터 타입 사용 (`*int`, `*float64`, `*string`)
- 서버 생성 타임스탬프: `RegDttm time.Time` (DB 컬럼 `reg_dttm`)
- **모든 "현재 시각"은 DB `now()`를 단일 진실의 원천으로 사용한다.** WAS 다중 인스턴스 시계 드리프트 방지가 목적.
  - INSERT 기본값: DDL `default now()` + struct 태그 `gorm:"->"` (읽기 전용, INSERT에서 제외)
  - map 기반 UPSERT: `gorm.Expr("NOW()")`
  - service/handler에서 현재 시각 변수 필요 시: `com.Now(ctx, sqlTx)` 호출
  - 예외: `time.Since(start)` 같은 상대 시간 측정(latency 계산)은 `time.Now()` 허용

## 아키텍처 및 SQL 가이드

코드 구조와 트랜잭션 모델: [docs/architecture.md](docs/architecture.md) — 3계층(handler/service/dao), DAO 페어링(`xxx_dao.go`+`xxx_dao.sql`), GORM↔`*sql.Tx` 공동 트랜잭션.

`.sql` 파일 작성 규칙: [docs/sql-style.md](docs/sql-style.md) — 소문자, 탭 들여쓰기, named query 마커, 컬럼 leading-comma.

신규 피처 작업 시 두 문서를 먼저 읽고 시작한다.

## 인증

기본은 Firebase ID Token (`Authorization: Bearer <token>`). 검증 결과 sub은 `c.GetString(com.CtxKeyMemberID)` 로 꺼낸다.

다른 발급자로 교체하려면 [com/middleware.go](com/middleware.go) 의 `InitFirebaseJWKS`, `FirebaseAuth` 만 수정한다.

## 응답 envelope

성공: `{"ok": true, ...}`
실패: `{"ok": false, "error": "메시지"}`

helper: `com.OK(c, gin.H{...})`, `com.OKWithCount(c, n)`, `com.BadRequest(c, msg)`, `com.RespondServiceError(c, err)`.

## 배포

- Serverless Framework + `provided.al2023` + ARM64
- VPC 내 Lambda → RDS PostgreSQL 접근 시 NAT Gateway 필수 (Firebase JWKS는 `googleapis.com`)
- 기본 region: ap-northeast-2 (`AWS_REGION` 환경변수로 오버라이드 가능)

스테이지별 `.env` 파일: `.env.dev`, `.env.prod`. 커밋 금지 (`.gitignore` 등록됨).

## 새 피처 추가 체크리스트

1. DB에 테이블 먼저 생성 (DDL → [docs/create-tables.md](docs/create-tables.md) 에 추가)
2. 새 패키지 디렉토리 (예: `member/`)
3. 도메인 모델 + DAO (`member_dao.go` + `member_dao.sql`) — 커스텀 쿼리가 없으면 GORM 자동 CRUD만 써도 됨
4. service (`member_service.go`) — 트랜잭션 + 비즈니스 로직
5. handler (`member_handler.go`) — HTTP 바인딩 + 응답
6. `main.go` 에 라우트 등록 (인증 필요 시 `authorized` 그룹 아래)
7. 필요 시 `test/<feature>/` 에 외부 패키지 테스트 추가

**주의: AutoMigrate 사용 금지. DB 테이블 구조는 직접 DDL로 관리한다.**
