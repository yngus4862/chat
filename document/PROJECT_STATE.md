# PROJECT_STATE.md

> **프로젝트:** 사내 메신저(카카오톡 유사) + 서버 구축 (온프레미스)  
> **대상 클라이언트:** Windows(C#) / Android(Java) / iOS(Swift)  
> **서버 배포:** Windows Docker 환경에서 Ubuntu 24.04 컨테이너로 운영  
> **최종 업데이트:** 2026-02-25 (Asia/Seoul) — Go 전환 확정 + DevContainer 기동 안정화 + 스모크 테스트 추가

---

## 1) 최신 상태 요약 (현재까지 확정/합의된 사항)

- 사용자 규모 20명(DAU/동접 20), 메시지 500건/일, 첨부 1GB/일(최대 300MB) 기준으로 설계 진행 중
- 서버는 **Windows Docker** 위에서 **Ubuntu 24.04 컨테이너**로 구성(Compose 기반 운영 전제)
- 개발환경은 **workspace/.devcontainer/** 아래에 인프라(Compose) 설정을 분리해 관리하며, **앱도 DevContainer(app)에서 실행(air 핫리로드)**
- Keycloak/MinIO 헬스체크 및 이미지 Pull(사내망 HTTPS Proxy) 이슈를 해결하여 인프라 컨테이너 pull·up을 확인
- 서버(REST/Realtime) 구현 언어는 **Go** 중심으로 진행(REST: Gin, Realtime: WebSocket)
- 데이터는 **PostgreSQL**, 캐시/이벤트는 **Redis**, 검색은 **Elastic**, 파일은 **MinIO(S3 호환)** 로 분리
- 인증/조직 연동은 **Keycloak(IdP)** 로 중앙화(OIDC), LDAP/SSO 혼용은 Keycloak에서 흡수하는 방향
- 읽음/미읽음은 **room_members.last_read_message_id** 기반(성능/구현 단순성 우선)으로 선택
- 링크 프리뷰는 서버에서 수집하되 **SSRF 방어(사설대역 차단/timeout/redirect 제한/최대바이트 제한)** 를 필수로 포함
- 관측성은 **Prometheus + Grafana(메트릭/알람)** + **Elastic(로그/검색)** 로 초기부터 포함
- 백업/복구 목표는 **RTO 1시간, RPO 15분**이며 Postgres WAL 기반 복구점 확보를 전제
- MVP는 9~12주(권고 10주) 로드맵으로 쪼개어 기능/운영을 함께 구축하는 계획

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
- **Backend:** Go (REST + WebSocket)
- **DB:** PostgreSQL
- **Cache/Event:** Redis (초기 pubsub + 캐시)
- **Search:** Elasticsearch
- **Object Storage:** MinIO (Presigned URL, Multipart/TUS 중 택1)
- **Auth/SSO:** Keycloak (OIDC)
- **Observability:** Prometheus + Grafana + Elastic(로그)

**선정 이유(요약):** 동시성/실시간 처리 적합 + 단일 바이너리 배포 용이 + 컨테이너/온프레미스 운영 단순(규모 대비 과잉 설계 방지)

---

## 4) 핵심 설계 결정 사항(Decisions)

### 4-1. 실시간/신뢰성
- 전달 방식: WebSocket(기본) + 오프라인/백그라운드는 푸시(모바일)
- 신뢰성: **at-least-once + idempotency(clientMsgId)** 로 중복 방지
- 순서 보장: **채팅방 단위 서버 시퀀싱(message_id 증가)**

### 4-2. 읽음/미읽음
- **선택:** (b) room_members.last_read_message_id / last_read_at 기반
- 이유: 쓰기 폭발 방지, 구현 단순, 20명 규모에서 충분한 UX 제공
- 확장 옵션: 감사/컴플라이언스 요구가 생기면 (a) read_receipts 하이브리드 추가

### 4-3. 파일/이미지
- 파일 본문은 DB 저장 금지(원칙), **MinIO**에 저장
- DB에는 메타(키/권한/참조/해시/바이러스 상태)만 저장
- 바이러스 스캔: 옵션이지만 권고(ClamAV Worker)
- 썸네일: Worker로 생성 후 thumb_key 저장

### 4-4. 링크 미리보기
- 서버에서 fetch + 캐시(일관 UX)
- SSRF 방어 필수: 사설대역/메타데이터 IP 차단, redirect 제한, timeout/바이트 제한

### 4-5. 푸시(플랫폼별)
- Android: FCM
- iOS: APNs
- Windows: MVP에서는 “오프라인 푸시” 제외(온라인 실시간 알림 + 토스트 중심). 필요 시 WNS 검토

---

## 5) 데이터 모델(요약)
- users, profiles, org_units, org_members
- chat_rooms, room_members, messages, message_mentions, pins
- attachments, links, link_previews
- polls, poll_options, poll_votes
- surveys, survey_questions, survey_choices, survey_responses, survey_answers
- push_tokens, notifications
- audit_logs
- (선택) read_receipts

---

## 6) 구현/산출물 진행 상태(Status)

### 6-1. 완료(설계)
- 아키텍처/스택 권고안 확정
- 읽음/미읽음 방식 선택
- 파일 저장/링크 프리뷰/SSRF 방어 원칙 정의
- 운영(관측성/백업 목표) 요구 반영

### 6-2. 완료(개발환경/인프라 기동)
- **workspace/.devcontainer/** 분리 구조 확정(인프라 설정/소스 분리)
- Compose 기반 인프라 스택 작성 및 기동 확인(nginx/postgres/redis/keycloak/minio + profiles: search/observability/workers)
- Keycloak: 커스텀 이미지 빌드 제거(패키지 매니저 이슈 회피), **/health/ready(9000)** 기반 healthcheck로 안정화, realm import 마운트 적용
- MinIO: 이미지 레퍼런스 오타(`::`) 수정, 버킷 생성 init job(옵션 profile) 구성
- 사내망에서 Docker 이미지 pull 실패(HTTPS proxy 미설정) → Docker Desktop **HTTPS Proxy 설정**으로 해결

- DevContainer 안정화(Windows+WSL2)
  - Go toolchain: `go not found` 재발 방지를 위해 GOPATH/PATH 명시 및 tool 설치를 vscode 유저로 수행
  - air: 모듈 경로 변경 대응(`github.com/cosmtrek/air` → `github.com/air-verse/air`) 반영
  - Windows 포트 정책: Redis(6379) 등 인프라 포트는 기본 publish를 피하고 `expose` 중심으로 구성(필요 시 localhost 바인딩 + 대체 포트 사용)
  - Git: Windows bind mount에서 발생하는 `dubious ownership` 이슈를 위해 `safe.directory` 예외 처리(권장)
  - VS Code Server: `/home/vscode/.cache/Microsoft` 생성 EACCES 가능성 대비(권장)

### 6-3. 다음 단계(실행 필요)
- Nginx 라우팅 최종 검증: `/auth`(Keycloak) + `/api` + `/realtime`(WebSocket) (앱 포트/경로 고정 후)
- DB 마이그레이션 파이프라인(자동 적용) 구축(**golang-migrate** 기준으로 고정)
- Keycloak Realm/Client/Role 초기화 자동화(JSON 확정 및 import 흐름 정리)
- MinIO presigned URL 흐름 + CORS/정책 점검(도메인 기준)
- Elastic 인덱스 템플릿/인덱서(outbox) 구성 + Kibana profile 운영(리소스 확인)
- Prometheus/Grafana(메트릭/대시보드/알람) 기본 세팅 + 앱/exporter 연결
- 백업/복구 스크립트 및 리허설(runbook)

---

## 7) Docker Compose 작업 목록(우선순위)

- [x] 1. 레포/폴더 구조 확정(infra/services/docs) → **workspace/.devcontainer/** 중심으로 정리
- [x] 2. `.env.template` 및 시크릿 정책 정의(DPAPI/Secrets) → **.env.example** 제공 + 로컬 `.env` 운영
- [ ] 3. 각 서비스 Dockerfile(DEV/PROD) 작성  
  - 현재는 공식 이미지 사용이 우선이며, 커스텀 빌드는 “필요 시”로 보류(특히 Keycloak)
- [x] 3-1. DevContainer 앱 이미지 빌드 안정화(오류 재발 방지)
  - GOPATH/PATH 명시, air 모듈 경로 변경 대응(air-verse), app 기동 안정화
- [x] 3-2. Windows 호스트 포트 publish 정책 정리
  - 인프라 포트는 기본 `expose`로만 사용(필요 시 localhost 바인딩 + 대체 포트)
- [x] 4. `docker-compose.yml` 스켈레톤 생성(nginx/postgres/redis/minio/keycloak + profiles: elastic/prom/grafana/workers)
- [x] 5. Nginx 설정(WebSocket/업로드 제한/timeout/TLS) 초안 반영
- [ ] 6. DB 마이그레이션 job(service) 추가(도구 선택 후 반영)
- [ ] 7. Keycloak realm import/export 자동화(Realm/Client/Role JSON 확정 후 고정)
- [x] 8. MinIO init job(버킷/정책) 추가(옵션 profile)
- [ ] 9. Elastic 템플릿 + 인덱서 워커 연결(outbox 구현 후)
- [x] 10. 관측성(메트릭/로그) 기본 배선(Prometheus/Grafana 스켈레톤)
- [ ] 11. 백업/복구(runbook + 스크립트) 추가
- [x] 12. 스모크 테스트 추가(REST + WS + DB 연동 자동 검증)
- [ ] 12-1. 부하 테스트(k6 등) 스크립트 추가
- [ ] 13. CI/CD(Jenkins) 골격 구성

---

## 8) 오픈 이슈(Open Questions / To Decide)

- 업로드 재개(Resumable): S3 Multipart vs tus 중 어느 쪽으로 고정할지
- Realtime Gateway를 Chat API와 동일 프로세스로 둘지(초기) / 분리할지(확장)
- 로그 수집 에이전트: Filebeat vs Vector 중 선택
- Windows 클라이언트 배포 형태(WPF/WinUI, 스토어 여부) 및 오프라인 푸시(WNS) 도입 시점
- 사내망 환경에서 Docker Desktop **HTTPS Proxy 설정 값/온보딩 절차**를 문서로 고정(신규 개발 PC 재현성)

---

## 9) MVP(10주 권고) 개요(요약)

- 1~2주: Auth/Org/Profile/검색 기반
- 3~5주: 방/메시지/읽음 + 재연결 동기화
- 6~7주: 파일/이미지/링크 프리뷰(보안 포함)
- 8주: Elastic 검색 + 인덱싱 파이프라인(outbox)
- 9주: 투표 + 설문 + 통계
- 10주: 모바일 푸시 + 운영 대시보드 + 백업/복구 리허설

---
