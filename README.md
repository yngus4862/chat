# 사내 메신저 서버 (Go) — DevContainer 개발환경 + 스켈레톤

> 이 README는 **Go 기반(REST + WebSocket)** 서버 스켈레톤과  
> **VS Code Dev Containers(Windows + WSL2 + Docker Desktop)** 환경에서 “바로 개발”할 수 있는 실행 방법을 정리합니다.

---

## 1) 구성 개요

- **API(REST)**: `:8080`
- **Realtime(WebSocket)**: `:8081`
- **DB**: PostgreSQL (docker compose)
- **Cache/Event**: Redis (docker compose)
- **Object Storage**: MinIO (docker compose)
- **Auth(옵션)**: Keycloak (docker compose profile)

앱 컨테이너는 기본적으로 `air`로 핫리로드합니다.

---

## 2) 빠른 시작 (VS Code Dev Containers)

### 전제
- Windows + WSL2(Ubuntu) + Docker Desktop
- VS Code 확장: **Dev Containers**

### 실행
1. VS Code로 프로젝트 폴더를 엽니다.
2. Command Palette → **Dev Containers: Reopen in Container**
3. 컨테이너가 뜨면 터미널에서 아래 확인:
   ```bash
   go version
   air -v
   ```

> DevContainer가 자동으로 compose를 올리고(app/postgres/redis/minio 등) app은 air로 실행됩니다.

---

## 3) 엔드포인트(MVP)

### REST
- `GET /healthz`
- `GET /readyz`
- `POST /v1/rooms`
- `GET /v1/rooms`
- `POST /v1/rooms/{roomId}/messages`
- `GET /v1/rooms/{roomId}/messages?limit=50`

기본 테스트:
```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

### WebSocket
- `GET ws://localhost:8081/ws?roomId=...`

간단 WS 테스트(예: wscat 사용 시):
```bash
# npm i -g wscat (호스트에 설치)
wscat -c "ws://localhost:8081/ws?roomId=1"
# {"content":"hello"}
```

---

## 4) 포트/접근 정책(중요)

Windows 환경에서 “유명 포트(예: 6379)”를 0.0.0.0로 publish하면
보안 정책/예약 포트/점유 때문에 실패하는 경우가 있습니다.

그래서 현재 compose는 인프라(Postgres/Redis/MinIO/Keycloak)를 **expose만** 하고,
호스트 publish는 최소화하는 구성을 권장합니다.

호스트에서 DB/Redis/MinIO에 직접 붙고 싶으면,
`.devcontainer/devcontainer.compose.yaml`에 `ports:`를 임시로 추가하세요.

예시(원하는 경우에만):
```yaml
redis:
  ports:
    - "127.0.0.1:16379:6379"
postgres:
  ports:
    - "127.0.0.1:15432:5432"
minio:
  ports:
    - "127.0.0.1:19000:9000"
    - "127.0.0.1:19090:9090"
```

---

## 5) 환경 변수

- `.env`는 커밋하지 않습니다.
- `.env.example`를 복사해서 `.env`로 사용합니다.

핵심(기본값 예시):
- `APP_PORT=8080`
- `APP_WS_PORT=8081`
- `POSTGRES_HOST=postgres`
- `REDIS_HOST=redis`
- `MINIO_ENDPOINT=minio:9000`

---

## 6) 트러블슈팅

### 6-1) Git: `dubious ownership in repository at '/workspace'`
Windows bind mount 환경에서 파일 소유자가 컨테이너 사용자와 다르게 보일 때 Git이 보안상 차단합니다.

컨테이너 안에서 아래 실행:
```bash
git config --global --add safe.directory /workspace
```

(권장) Dockerfile에 system-level safe.directory를 넣어두면 매번 안 해도 됩니다.

### 6-2) VS Code Server: `EACCES: permission denied, mkdir '/home/vscode/.cache/Microsoft'`
`/home/vscode/.cache` 하위 권한이 꼬이면 발생합니다.

컨테이너 안에서 임시 해결:
```bash
mkdir -p /home/vscode/.cache/Microsoft
sudo chown -R vscode:vscode /home/vscode/.cache
```

(권장) Dockerfile에서 `.cache/Microsoft`를 미리 생성하고 chown 처리로 고정하세요.

### 6-3) Windows 로그: `AttachConsole failed`
이건 보통 **Windows 호스트 쪽(ConPTY/node-pty)** 경고로,
컨테이너가 정상 실행되고 VS Code Server가 올라가면 치명적 이슈가 아닙니다.

### 6-4) 사내망에서 Docker pull 실패(프록시/DNS)
- Docker Desktop → **HTTPS Proxy 설정**
- WSL `/etc/resolv.conf` 확인
- 필요 시 compose에 `dns:` 지정

---

## 7) 스모크 테스트(자동)

DevContainer 접속 후 아래 한 줄로 REST + WebSocket + DB 연동까지 기본 검증을 수행합니다.

```bash
make smoke
# 또는
go run ./cmd/smoketest
```

CI/자동화용(서버가 떠 있어야 함):
```bash
go test -tags=integration ./tests -v
```

---

## 8) 다음 확장 로드맵(요약)
- Nginx Reverse Proxy 경로 고정(`/api`, `/realtime`, `/auth`)
- 마이그레이션(golang-migrate) job/자동 적용
- outbox → indexer → Elastic 검색 파이프라인
- 파일 업로드 presigned/multipart + 썸네일 워커
- 관측성(Prom/Grafana/로그) 정식 대시보드/알람 세트