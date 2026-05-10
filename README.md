# go-gin-lambda-starter

Go + Gin + Serverless Framework + AWS Lambda 백엔드 스타터킷.

## 기술 스택

| 항목 | 기술 |
|------|------|
| Language | Go 1.26 |
| Web Framework | Gin |
| ORM | GORM + PostgreSQL driver |
| Database | PostgreSQL (Amazon Aurora 권장) |
| Auth | Firebase ID Token (JWT 검증, 교체 가능) |
| Deploy | Serverless Framework → AWS Lambda (`provided.al2023`, ARM64) + API Gateway v2 |
| Region | ap-northeast-2 (기본값, 변경 가능) |

## 사전 요구사항

- Go 1.26+
- Node.js 18+ (Serverless Framework 실행)
- PostgreSQL (로컬 또는 원격)
- Firebase 프로젝트 (또는 다른 OIDC 발급자)

## 개발 환경 초기 설정

처음 클론한 뒤 한 번 실행. Go 개발 도구 설치 + git 훅 활성화.

```bash
npm install
npm run setup
```

- `goimports`, `golangci-lint` 설치
- `.githooks/` 활성화 (pre-commit: 자동 포맷, pre-push: 린트)

## 로컬 개발

```bash
cp .env.example .env.dev
# .env.dev 에 DB 접속 정보 + FIREBASE_PROJECT_ID 입력

# Linux/Mac
set -a; source .env.dev; set +a; npm run local

# Windows PowerShell
Get-Content .env.dev | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') { $env:($matches[1].Trim()) = $matches[2].Trim() }
}
npm run local
```

서버는 `http://localhost:8080`에서 기동된다.

```bash
curl http://localhost:8080/health
# {"ok":true}
```

## 빌드 & 배포

```bash
# Lambda 바이너리 빌드 (linux/arm64)
npm run build

# dev 환경 배포
npm run deploy:dev

# prod 환경 배포
npm run deploy:prod
```

`.env.dev` / `.env.prod` 파일을 Serverless Framework가 자동 로드한다.
`SERVICE_NAME` 환경변수로 서비스명을 오버라이드할 수 있다.

## AWS VPC 필수 조건 (선택)

Lambda를 RDS Private Subnet에 배치하려면 `serverless.yml`의 `vpc` 블록을 활성화한다.
**VPC 내 Lambda는 NAT Gateway 없이는 인터넷에 접근하지 못한다.** Firebase JWKS는
`googleapis.com`을 호출하므로 NAT Gateway가 없으면 cold start에서 즉시 종료된다.

NAT Gateway 구성 (AWS 콘솔):
1. VPC → Elastic IPs → Elastic IP 할당
2. VPC → NAT Gateways → NAT Gateway 생성 (Public Subnet에 배치)
3. Lambda 배치 Private Subnet의 라우트 테이블: `0.0.0.0/0 → NAT Gateway`

## CloudWatch 로그 조회

```bash
aws logs tail /aws/lambda/<service-name>-dev-api --follow --region ap-northeast-2
```

Git Bash에서 경로 변환 방지: `MSYS_NO_PATHCONV=1` 접두.

## 인증 엔드포인트

기본 미들웨어는 Firebase ID Token을 검증한다 (`Authorization: Bearer <token>`).
다른 JWT 발급자로 교체하려면 [com/middleware.go](com/middleware.go) 의 `InitFirebaseJWKS`,
`FirebaseAuth` 만 수정하면 된다.

| Method | Path | Auth | 설명 |
|--------|------|------|------|
| GET | /health | 불필요 | 헬스 체크 |
| GET | /example/items | 필요 | 회원의 모든 아이템 |
| GET | /example/items/:id | 필요 | 단건 조회 |
| POST | /example/items | 필요 | 신규 등록 |

## 프로젝트 구조

```
main.go              # 진입점 (Lambda/로컬 전환, 라우트 등록)
com/                 # 공통 인프라 — 새 프로젝트에서도 그대로 재사용
  config.go          # 환경변수 설정
  database.go        # GORM 초기화 (Lambda 풀 튜닝)
  middleware.go      # FirebaseAuth, Recovery, ApiLogging, RespondServiceError
  models.go          # ApiLog 모델 (다른 모델은 도메인 패키지에 추가)
  dbtx.go            # GORM ↔ *sql.Tx 브리지
  named_query.go     # SQL 템플릿 엔진
  time_dao.go(.sql)  # DB-side NOW() helper
  response.go        # OK/BadRequest helper
example/             # 샘플 3-tier 모듈 (복사해서 새 모듈 출발점으로)
  example_handler.go # HTTP 파싱 + 응답
  example_service.go # 트랜잭션 + 비즈니스 로직
  example_dao.go     # 커스텀 SQL (raw *sql.Tx)
  example_dao.sql    # named query 모음
docs/                # 컨벤션 가이드 — 새 피처 작업 전 읽기
test/                # 외부 패키지 테스트 (test/<pkg> 미러)
scripts/             # 빌드 스크립트
.githooks/           # pre-commit (포맷), pre-push (린트)
```

자세한 구조와 컨벤션은 [CLAUDE.md](CLAUDE.md) 와 [docs/](docs/) 참고.

## 새 모듈 추가 5단계

1. DB에 테이블 DDL 실행 ([docs/create-tables.md](docs/create-tables.md) 참고)
2. 새 패키지 디렉토리 생성 (예: `member/`)
3. `member_dao.go` + `member_dao.sql` (필요 시), `member_service.go`, `member_handler.go` 작성
4. GORM 모델이 필요하면 같은 패키지에 추가, 공통 모델만 `com/models.go` 에
5. `main.go`에 라우트 등록

## 코드 품질 (커밋 전)

```bash
npm run check  # 포맷 + 린트
```

## 라이선스

원하는 라이선스로 변경.
