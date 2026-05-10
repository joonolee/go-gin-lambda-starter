# 서버 코드 구조 및 트랜잭션 모델

## 1. 계층 구조 (3-tier)

```
handler (<feature>_handler.go)
  └─ service (<feature>_service.go)
       └─ dao (<feature>_dao.go + <feature>_dao.sql)
```

- **handler**: HTTP 바인딩/응답 직렬화. 비즈니스 로직 금지.
- **service**: 트랜잭션 경계, 도메인 로직, GORM 자동 CRUD (`Create`/`Updates`/`First` 등).
- **dao**: GORM으로 표현하기 번거로운 커스텀 쿼리 담당 (`ON CONFLICT ... RETURNING`, JOIN/CTE/집계 등).

## 2. GORM ↔ `*sql.Tx` 공동 트랜잭션

service는 `db.Transaction(func(tx *gorm.DB) error {...})`로 트랜잭션을 시작한다.
내부에서 `com.SQLTx(tx)`로 `*sql.Tx`를 추출해 DAO에 전달한다.
자동 CRUD는 GORM `tx`로, 커스텀 쿼리는 `*sql.Tx`로 — 동일 물리 트랜잭션.

```go
err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    sqlTx, err := com.SQLTx(tx)   // com/dbtx.go
    if err != nil { return err }
    // DAO 호출: sqlTx 사용
    // GORM CRUD: tx 사용
    return nil
})
```

`com.SQLTx(tx)` 추출은 항상 한 곳에서만 한다 (`com/dbtx.go`).

## 3. DAO 파일 페어링

`xxx_dao.go` + `xxx_dao.sql`을 같은 디렉터리에 1:1로 둔다.
한 `.sql` 파일에 여러 named query를 `-- name: <key>` 마커로 구분한다.
`.sql`은 `//go:embed` 디렉티브로 컴파일 타임에 바이너리에 포함된다.
SQL 작성 규칙은 [docs/sql-style.md](sql-style.md) 참고.

## 4. 공통 DAO 위치

`com/<name>_dao.go` + `com/<name>_dao.sql`.
도메인을 넘어 공유되는 헬퍼 쿼리만 들어간다. 현재: `time_dao` (DB 현재 시각).

## 5. 피처 DAO 위치

`<feature>/<feature>_dao.go` + `<feature>/<feature>_dao.sql`.
한 피처 패키지에 한 세트(handler/service/dao+sql). DAO 없이 GORM 자동 CRUD만으로 충분하면 생략 가능.
샘플: `example/example_dao.{go,sql}`.

## 6. CRUD 분담 — GORM vs DAO

| 종류 | 위치 |
|---|---|
| 단순 INSERT/UPDATE/SELECT by id | service에서 `tx.Create` / `tx.Updates` / `tx.First` 직접 |
| `ON CONFLICT DO NOTHING RETURNING` 등 raw INSERT | `xxx_dao.sql` + `*sql.Tx` |
| JOIN/CTE/IN/집계 SELECT | `xxx_dao.sql` + `*sql.Tx` |

DAO 파일은 커스텀 쿼리만 담는다. `tx.First(&model, id)` 류는 service에 그대로 둔다.

## 7. 시간 처리

모든 "현재 시각"은 DB `now()`를 단일 진실의 원천으로 사용한다. WAS 다중 인스턴스 시계 드리프트 방지가 목적.

- **INSERT 기본값**: DDL `default now()` + GORM struct 태그 `gorm:"->"` (읽기 전용, INSERT에서 제외)
- **map 기반 UPSERT**: `gorm.Expr("NOW()")`
- **service/handler에서 현재 시각 변수 필요 시**: `com.Now(ctx, sqlTx)` 호출
- **예외**: `time.Since(start)` 같은 상대 시간 측정(latency 계산)은 `time.Now()` 허용

## 8. 응답 envelope

표준 응답 형식.

- 성공: `{"ok": true, ...}`
- 실패: `{"ok": false, "error": "메시지"}`

`com/response.go` helper:
- `com.OK(c, gin.H{...})` — 200 + `{"ok": true, ...}`
- `com.OKWithCount(c, n)` — 200 + `{"ok": true, "count": n}`
- `com.BadRequest(c, msg)` — 400
- `com.RespondServiceError(c, err)` — 500 (내부 에러는 로그만, 응답엔 일반 메시지)

## 9. 테스트 디렉터리

모든 `*_test.go`는 소스와 같은 디렉터리에 두지 않고 **프로젝트 루트의 `test/` 아래**에 패키지 미러로 둔다.

```
com/named_query.go      ↔  test/com/named_query_test.go
example/example_dao.go  ↔  test/example/example_dao_test.go
```

테스트는 `package <pkg>_test` 외부 패키지 형식으로 작성해 exported API만 사용.
실행: `go test ./test/...` (또는 `npm test`).

## 10. 신규 피처 추가 체크리스트

1. DB에 테이블 DDL 실행 (`docs/create-tables.md` 에 추가)
2. `<feature>/<feature>_handler.go` — 라우팅/응답
3. `<feature>/<feature>_service.go` — 트랜잭션 + 비즈니스
4. `<feature>/<feature>_dao.go` + `<feature>_dao.sql` — 커스텀 쿼리 (없으면 생략)
5. 도메인 모델은 같은 패키지에 (공통 모델만 `com/models.go`)
6. `main.go`에 라우트 등록
7. `test/<feature>/...` 아래에 테스트 추가 (필요 시)
