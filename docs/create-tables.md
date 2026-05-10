# PostgreSQL CREATE TABLE DDL

스타터킷 기본 테이블 + 새 피처 추가 시 여기에 DDL 누적.

## 기본 테이블

```sql
-- 1. api_log : ApiLogging 미들웨어가 자동 기록하는 모든 요청 로그
create table api_log (
    id           bigserial    primary key,
    member_id    varchar(128),
    method       varchar(10)  not null,
    endpoint     varchar(200) not null,
    status_code  integer      not null,
    record_count integer,
    latency_ms   integer      not null,
    reg_dttm     timestamptz  not null default now()
);
create index idx_api_log_endpoint_dttm on api_log (endpoint, reg_dttm);

-- 2. item : example 모듈 샘플 테이블
create table item (
    id         bigserial    primary key,
    member_id  varchar(128) not null,
    name       varchar(100) not null,
    reg_pgm    varchar(500) not null,
    reg_dttm   timestamptz  not null default now()
);
create index idx_item_member on item (member_id, id desc);
```

## 컨벤션

- 모든 테이블에 `reg_dttm timestamptz not null default now()`
- 단일 PK는 `bigserial primary key` 또는 `bigint primary key` (앱 측 ID 받을 때)
- nullable 컬럼은 명시적으로 `null` 허용 (Go 모델은 포인터)
- 인덱스 이름: `idx_<table>_<columns>`
- UNIQUE 제약: `unique (col1, col2)`
- 멱등성 INSERT 패턴: `unique` 제약 + `on conflict (...) do nothing returning id`

## 새 피처 추가 시

피처 패키지 추가 후 이 파일 하단에 새 섹션으로 DDL 추가.
