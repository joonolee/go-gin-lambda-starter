---
description: dao .sql 파일의 sql-style.md 규칙 위반 + Go↔SQL named query 정합성 + go:embed 디렉티브 누락을 검사한다
---

# check-sql

`go-gin-lambda-starter`의 SQL/DAO 정합성을 정적 검증한다. 권위 출처는 [docs/sql-style.md](../../docs/sql-style.md).

검사 대상은 리포지토리 전체의 `*_dao.sql` 와 페어링되는 `*_dao.go` 파일이다.

## 수행 절차

다음 3개 검사를 순서대로 수행하고, 각각의 결과를 분리된 섹션으로 보고한다. 위반이 0건이면 해당 섹션을 "OK"로 표시한다.

### 1. `docs/sql-style.md` 규칙 위반 검사

대상: `**/*_dao.sql` 전부.

각 파일을 Read한 뒤 라인 단위로 다음 위반을 검출한다.

**1-A. 대문자 SQL 키워드** — 토큰이 단어 경계로 분리된 형태로 등장하면 위반.

검사할 키워드 (대소문자 구분, 정확히 이 토큰이 코드 영역에 나타나면 위반):
`SELECT`, `FROM`, `WHERE`, `AND`, `OR`, `WITH`, `AS`, `UPDATE`, `SET`, `INSERT`, `INTO`, `VALUES`, `ON`, `CONFLICT`, `DO`, `NOTHING`, `RETURNING`, `COALESCE`, `COUNT`, `NOW`, `ANY`, `CASE`, `WHEN`, `THEN`, `ELSE`, `END`, `ORDER`, `BY`, `LIMIT`, `OFFSET`, `GROUP`, `HAVING`, `JOIN`, `LEFT`, `RIGHT`, `INNER`, `OUTER`, `IS`, `NULL`, `NOT`, `IN`, `EXISTS`, `DISTINCT`, `UNION`, `ALL`

**제외**:
- `/* ... */` 블록 주석 내부 (한글 설명에 영문 대문자가 섞일 수 있음)
- `'...'` 문자열 리터럴 내부
- `-- name:` 마커 라인의 키 부분은 snake_case라 해당 없음

Grep으로 후보를 추리고, 후보 라인에 대해서만 Read로 컨텍스트를 확인하면 효율적이다.

**1-B. 스페이스 들여쓰기** — 들여쓰기는 탭 1개 단위(sql-style.md §4). 라인이 공백 문자(스페이스)로 시작하면 위반.

정규식: `^ +` (스페이스로 시작하는 모든 라인). 빈 라인은 제외.

**1-C. Trailing comma** — 컬럼 나열은 leading-comma 규칙(sql-style.md §5). 라인이 `,` 또는 `,` 뒤 공백/주석만으로 끝나면 위반.

정규식: `,\s*$` 또는 `,\s*(/\*.*\*/)?\s*$`. 단, 다음은 제외:
- `(` 직전이 아닌 함수 인자 목록의 마지막이 아닌 경우 (e.g. `coalesce(a, b)`는 정상)
- `values` 절 내부의 일반 `,` 분리 (별도 라인이 아니면 무시)

판정 단순화: **라인 끝이 `,`로 끝나는 경우만 위반**으로 본다. (leading-comma 규칙상 컬럼 분리 콤마는 다음 라인의 맨 앞에 와야 하므로.)

**보고 형식** (위반 1건당 1줄):

```
[STYLE] <파일경로>:<라인번호> <위반 종류> — <위반 내용 발췌>
```

예:
```
[STYLE] example/example_dao.sql:7 uppercase keyword — "FROM item"
[STYLE] example/example_dao.sql:5 space indent — "    ,name"
[STYLE] example/example_dao.sql:6 trailing comma — ",reg_dttm,"
```

### 2. Named Query 키 양방향 매칭

각 `_dao.sql` ↔ 페어링 `_dao.go` 쌍에 대해:

**SQL 측 키 추출**: 정규식 `^\s*--\s*name:\s*(\S+)` 으로 마커의 key를 수집.

**Go 측 키 추출**: 정규식 `MustGet\("([^"]+)"\)` 으로 호출 키를 수집. (현재 코드베이스가 사용하는 호출 형태. 다른 헬퍼가 추가되면 확장하되, 지금은 이 패턴만으로 충분.)

**비교**:
- SQL에 정의되어 있으나 Go에서 한 번도 호출하지 않는 키 → **orphan SQL**
- Go에서 호출하는데 SQL에 정의가 없는 키 → **dangling Go call**

페어링 규칙: `<dir>/<base>_dao.sql` ↔ `<dir>/<base>_dao.go` (동일 디렉터리, 동일 베이스네임).

**보고 형식**:

```
[NQ-ORPHAN]   <sql 파일경로>: <키>   (SQL에만 정의됨, Go 호출 없음)
[NQ-DANGLING] <go 파일경로>:<라인번호> <키>   (Go에서 호출, SQL 정의 없음)
```

### 3. `//go:embed` 디렉티브 누락 검사

대상: `*_dao.sql`이 존재하는 모든 디렉터리.

페어링되는 `_dao.go`를 Read해서 다음 두 조건을 만족하는지 확인:
- `_ "embed"` 임포트 존재
- `//go:embed <base>_dao.sql` 디렉티브 존재 (베이스네임이 페어링되는 SQL과 일치해야 함)

둘 중 하나라도 빠지면 위반.

**보고 형식**:

```
[EMBED] <go 파일경로>: missing `//go:embed <expected>_dao.sql` directive
[EMBED] <go 파일경로>: missing `_ "embed"` import
[EMBED] <sql 파일경로>: no paired *_dao.go found
```

## 최종 출력 형식

```
=== check-sql 결과 ===

[1] SQL 스타일 (docs/sql-style.md)
<위반 라인들 또는 "OK">

[2] Named query 키 정합성
<위반 라인들 또는 "OK">

[3] //go:embed 디렉티브
<위반 라인들 또는 "OK">

요약: 스타일 N건, 정합성 M건, embed K건
```

총 위반 0건이면 마지막에 "모든 검사 통과" 한 줄 추가.

## 구현 노트

- 파일 탐색은 Glob (`**/*_dao.sql`, `**/*_dao.go`), 패턴 매칭은 Grep을 우선 사용. Read는 컨텍스트 확인이 필요할 때만.
- 한 번에 여러 Grep을 병렬로 실행해 효율을 높인다.
- 코드 수정은 하지 않는다. **읽기 전용 검증 커맨드**다. 위반이 발견되어도 자동 수정하지 않고 사용자에게 보고만 한다.
- false positive를 줄이기 위해 주석/문자열 내부는 검사에서 제외한다. 판정이 애매한 경우 위반으로 보고하되 `(uncertain)` 표시를 덧붙인다.
