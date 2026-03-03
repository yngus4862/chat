# PROJECT_STATE.md

> **프로젝트:** 사내 메신저(카카오워크/카카오톡 유사) + 서버 구축(온프레미스)  
> **대상 클라이언트:** Windows(C#) / Android(Java) / iOS(Swift)  
> **서버 배포:** Windows Docker 환경에서 Linux 컨테이너(Compose)로 운영  
> **최종 업데이트:** 2026-03-03 (Asia/Seoul) — Go 기반 메시징 서버 MVP(REST+WS+DB+Redis) + 서비스 제어 + 스모크 테스트 반영

---

## 1) 최신 상태 요약 (현재까지 확정/합의된 사항)

- 사용자 규모 20명(DAU/동접 20), 메시지 500건/일, 첨부 1GB/일(최대 300MB) 기준으로 설계 진행 중
- 서버는 **Windows Docker Desktop** 위에서 **Linux 컨테이너**로 구성(Compose 기반 운영 전제)
- 개발환경은 **workspace/.devcontainer/** 아래에 인프라(Compose) 설정을 분리해 관리하며, **앱도 DevContainer(app)에서 실행(air 핫리로드)**
- 서버 구현 언어는 **Go**로 확정(REST: Gin, Realtime: WebSocket)
- 데이터는 **PostgreSQL**, 캐시/이벤트/브로드캐스트는 **Redis(pubsub/cache)**, 파일은 **MinIO(S3 호환)** 로 분리
- 인증/조직 연동은 **Keycloak(IdP, OIDC)** 로 중앙화(옵션 profile로 토글 가능)
- 읽음/미읽음은 **room_members.last_read_message_id** 기반(성능/구현 단순성 우선)으로 선택
- 링크 프리뷰는 서버에서 수집하되 **SSRF 방어(사설대역 차단/timeout/redirect 제한/최대바이트 제한)** 를 필수로 포함
- 관측성은 **Prometheus + Grafana(메트릭/알람)** + **Elastic(로그/검색)** 로 초기부터 포함(프로파일 기반)
- 운영 편의를 위해 서버 프로세스는 **종료/재시작/상태조회**(콘솔 + Admin API + CLI) 기능을 갖춘다

---

## 2) 시스템 아키텍처(텍스트)

```
[Win C#]   [Android]   [iOS]
   | REST/HTTPS   | WSS(WebSocket)   | Push(APNs/FCM)
   v              v                  v
            [Nginx Reverse Proxy(TLS)]
                    |
  ┌─────────────────┼───────────────────────┐
  v                 v                       v
[Chat API]     [Realtime Gateway]     [Notification Svc]
(Go REST)       (Go WS Hub)            (Push + Badge)
  |                 |                       |
  ├───────┬─────────┘                       |
  v       v                                 v
[Postgres] [Redis(pubsub/cache)]     [Push Providers]
  |
  ├── Outbox/Event → [Indexer Worker] → [Elastic]
  |
  └── File metadata → [MinIO(S3)] (+ optional ClamAV)
           |
     Thumbnails/Preview Worker
```

---

## 3) 기술 스택(권고안) 및 근거

### 권고 스택(A)
- **Backend:** Go (Gin + WebSocket)
- **DB:** PostgreSQL
- **Cache/Event:** Redis (초기 pubsub + 캐시)
- **Search:** Elasticsearch (옵션 profile)
- **Object Storage:** MinIO (Presigned URL, Multipart/TUS 중 택1)
- **Auth/SSO:** Keycloak (OIDC, 옵션 profile)
- **Observability:** Prometheus + Grafana + Elastic(로그, 옵션 profile)

**선정 이유(요약):** 동시성/실시간 처리 적합 + 단일 바이너리 배포 용이 + 컨테이너/온프레미스 운영 단순(규모 대비 과잉 설계 방지)

---

## 4) 핵심 설계 결정 사항(Decisions)

### 4-1. 실시간/신뢰성
- 전달 방식: WebSocket(기본) + 오프라인/백그라운드는 푸시(모바일)
- 신뢰성: **at-least-once + idempotency(clientMsgId)** 로 중복 방지
- 순서 보장: **채팅방 단위 서버 시퀀싱(message_id 증가)**

### 4-2. 읽음/미읽음
- 선택: room_members.last_read_message_id / last_read_at 기반
- 이유: 쓰기 폭발 방지, 구현 단순, 20명 규모에서 충분한 UX 제공

### 4-3. 파일/이미지
- 파일 본문은 DB 저장 금지(원칙), **MinIO**에 저장
- DB에는 메타(키/권한/참조/해시/바이러스 상태)만 저장

### 4-4. 링크 미리보기
- 서버에서 fetch + 캐시(일관 UX)
- SSRF 방어 필수: 사설대역/메타데이터 IP 차단, redirect 제한, timeout/바이트 제한

### 4-5. 푸시(플랫폼별)
- Android: FCM
- iOS: APNs
- Windows: MVP에서는 “오프라인 푸시” 제외(온라인 실시간 알림 + 토스트 중심). 필요 시 WNS 검토

### 4-6. 서비스 제어(운영 편의)
- 서버 프로세스는 “서비스처럼” 동작해야 하며, **종료/재시작/상태조회**를 제공한다.
- 제어 인터페이스는 3종을 제공한다.
  - (1) **콘솔 명령(stdin)**: `status | stop | restart | help`
  - (2) **관리 API(HTTP)**: `GET /admin/status`, `POST /admin/stop`, `POST /admin/restart`
  - (3) **CLI(chatctl)**: `chatctl status|stop|restart`
- 관리 API는 기본적으로 **토큰 기반 인증(Bearer Token)** 을 필수로 하여 무단 종료/재시작을 방지한다.
- 운영 기본값: Admin API는 `127.0.0.1` 바인딩 + `ADMIN_TOKEN`이 없으면 비활성(안전 기본값)

---

## 5) 데이터 모델(요약)
- users, profiles, org_units, org_members
- chat_rooms, room_members, messages, message_mentions, pins
- attachments, links, link_previews
- push_tokens, notifications
- audit_logs
- (선택) read_receipts

---

## 6) 구현/산출물 진행 상태(Status)

### 6-1. 완료(설계)
- 아키텍처/스택 권고안 확정(Go 기준)
- 읽음/미읽음 방식 선택
- 파일 저장/링크 프리뷰/SSRF 방어 원칙 정의
- 운영(관측성/백업 목표) 요구 반영

### 6-2. 완료(개발환경/인프라 기동)
- **workspace/.devcontainer/** 분리 구조 확정(인프라 설정/소스 분리)
- Compose 기반 인프라 스택 작성 및 기동 확인(postgres/redis/keycloak/minio, profiles: search/observability)
- DevContainer 안정화(Windows+WSL2)
  - Go toolchain 경로(GOPATH/PATH) 및 tool 설치(air/migrate/dlv) 안정화
  - air: 모듈 경로 변경 대응(`github.com/cosmtrek/air` → `github.com/air-verse/air`) 반영
  - Windows 포트 정책: 인프라 포트는 기본 publish를 피하고 `expose` 중심으로 구성(필요 시 localhost 바인딩 + 대체 포트)
  - Git: Windows bind mount에서 `dubious ownership` 발생 시 `safe.directory` 예외 처리(권장)

### 6-3. 완료(서버 MVP 코드 — 레포 반영 필요)
- REST API(MVP): rooms/messages + health/ready 구현
- WebSocket(MVP): room join + broadcast 구현(단일 인스턴스 + Redis pubsub 확장)
- 마이그레이션 SQL 제공(migrations/0001_init.*.sql)
- 서비스 제어(Admin API/콘솔/CLI) 구현
- 스모크 테스트(cmd/smoketest) 구현

### 6-4. 다음 단계(실행 필요)
- Nginx 라우팅 최종 검증: `/auth`(Keycloak) + `/api` + `/realtime`(WebSocket)
- DB 마이그레이션 job(service) 추가(자동 적용 방식 확정) — golang-migrate 기준
- Keycloak Realm/Client/Role 초기화 자동화(JSON 확정 및 import 흐름 정리)
- MinIO presigned URL 흐름 + CORS/정책 점검(도메인 기준)

---

## 7) Docker Compose 작업 목록(우선순위)

- [x] 1. 레포/폴더 구조 확정 → workspace/.devcontainer 중심 정리
- [x] 2. .env.example 제공 + 로컬 .env 운영 정책
- [ ] 3. 각 서비스 Dockerfile(DEV/PROD) 작성(필요 시)
- [x] 4. infra compose 스켈레톤(postgres/redis/minio/keycloak + profiles) 기동 확인
- [x] 5. Windows 호스트 포트 publish 정책 정리(expose 기본)
- [ ] 6. DB 마이그레이션 job(service) 추가(자동 적용)
- [x] 6-1. 서비스 제어(Admin API/콘솔/CLI) 추가 + 운영 정책 문서화
- [x] 7. 스모크 테스트 추가(REST + WS + DB 연동 자동 검증)
- [ ] 8. Nginx 최종 라우팅/timeout/업로드 제한 고정
- [ ] 9. Elastic 템플릿 + 인덱서 워커(outbox) 연결(후속)

---

## 8) 오픈 이슈(Open Questions / To Decide)
- 업로드 재개(Resumable): S3 Multipart vs tus 중 고정
- Realtime Gateway를 Chat API와 동일 프로세스로 둘지(초기) / 분리할지(확장)
- 사내망 환경에서 Docker Desktop HTTPS Proxy 설정 온보딩 절차 문서 고정
