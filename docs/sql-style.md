# SQL 작성 규칙

`.sql` 파일 작성 컨벤션 단일 진실의 원천.

## 1. 파일/페어링

DAO Go 파일과 SQL 파일은 동일 디렉터리·동일 베이스네임으로 페어링한다.
한 `.sql` 파일에 여러 named query를 둔다.

```
example/example_dao.go ↔ example/example_dao.sql
com/time_dao.go        ↔ com/time_dao.sql
```

## 2. Named Query 마커

블록 시작은 `-- name: <snake_case_key>` 한 줄. 그 아래 첫 줄에 `/* <key> : <한글 설명> */` 주석.
다음 마커 또는 EOF까지가 본문.

```sql
-- name: select_item_by_id
/* select_item_by_id : id로 단건 조회 */
select
	id
	,name
from item
where id = :id
```

## 3. 소문자 원칙

모든 SQL 키워드와 내장함수는 **소문자**:
`select`, `from`, `where`, `and`, `or`, `with`, `as`, `update`, `set`, `insert into`, `values`,
`on conflict`, `do nothing`, `returning`, `coalesce`, `count`, `now`, `any`,
`case`/`when`/`then`/`else`/`end` 등.
식별자(테이블/컬럼/별칭)도 소문자 snake_case.

## 4. 들여쓰기

탭 1개 단위. 줄 끝 세미콜론 없음.

## 5. `select` 절

키워드는 자체 줄. 컬럼은 한 줄에 하나. 두 번째부터 **앞에 콤마**.

```sql
select
	id
	,member_id
	,name
from item
```

## 6. `where` 절

첫 조건은 `where <조건>`. 추가 조건은 `and`/`or`를 앞에 두고 한 단계 들여쓰기.

```sql
where member_id = $1
	and id = any($2)
```

## 7. CTE / `with`

괄호를 자기 줄에 두고 내부 본문 한 단계 들여쓰기.

```sql
with v_last
as
(
	select
		col1
	from t_xxx
	where member_id = $1
)
select
	col1
from v_last
```

## 8. 플레이스홀더

`pgx` 드라이버 기준 `$1`, `$2` 위치 인자.
IN 절은 `= any($N)` + Go 측 `[]string`/`[]int64` 슬라이스 직접 바인딩 (`pq.Array(...)` 불필요).

Named param (`:name` → `$N` 치환)은 `com.Queries.Render(params)` 를 통해 자동 처리된다.

## 9. 주석

블록 헤더 외 인라인 주석은 `/* ... */` 사용.
`--` 한 줄 주석은 **named query 마커 전용**으로 예약 (파서가 `^\s*--\s*name:` 마커를 잡으므로 다른 용도에 쓰면 오파스).

## 10. 금지

- 컬럼 alias에 따옴표/대문자 사용 금지
- 함수 호출에 대문자 사용 금지
- `select *` 지양 (필요한 컬럼만 명시)
- 하드코딩 `LIMIT` 상수보다 인자 권장

## 참조 예시

- [example/example_dao.sql](../example/example_dao.sql)
- [com/time_dao.sql](../com/time_dao.sql)
