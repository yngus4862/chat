# Go 메시징 서버 (REST + WebSocket)

## 구성
- REST API: `:8080`
- WebSocket: `:8081` (`/ws?roomId=...`)
- PostgreSQL: 메시지/방 저장
- Redis: WS broadcast(pub/sub) 확장
- Admin API(옵션): `:9099` (`/admin/status|stop|restart`)

## 빠른 실행(DevContainer)
1) Dev Containers: Reopen in Container
2) (처음 1회) 마이그레이션 적용
```bash
migrate -path ./migrations -database "postgres://appuser:appsecret@postgres:5432/chatapp?sslmode=disable" up
```
3) 서버 실행
```bash
air --config /workspace/.air.toml
# 또는
go run ./cmd/chatd
```
4) 스모크 테스트
```bash
make smoke
# 또는
go run ./cmd/smoketest -api http://127.0.0.1:8080 -ws ws://127.0.0.1:8081/ws
```

## API
### Health
- `GET /healthz` -> `{"status":"ok"}`
- `GET /readyz`  -> DB/Redis readiness

### Rooms
- `POST /v1/rooms` `{ "name": "room" }`
- `GET /v1/rooms?limit=50`

### Messages
- `POST /v1/rooms/{roomId}/messages` `{ "content": "hi", "clientMsgId": "..." }`
- `GET /v1/rooms/{roomId}/messages?cursor=...&limit=50`

### WebSocket
- `GET ws://localhost:8081/ws?roomId=1`
- send JSON: `{ "content":"hello", "clientMsgId":"..." }`
- receive JSON: Message object `{id, roomId, content, createdAt, ...}`

## 서비스 제어(Admin API)
- `ADMIN_TOKEN`이 **설정된 경우에만** Admin 서버가 실행됩니다.
- 기본 바인딩은 `127.0.0.1:9099` 권장(외부 노출 금지)

```bash
export ADMIN_TOKEN=change-me-long-random
curl -H "Authorization: Bearer ${ADMIN_TOKEN}" http://127.0.0.1:9099/admin/status
curl -XPOST -H "Authorization: Bearer ${ADMIN_TOKEN}" http://127.0.0.1:9099/admin/restart
```

또는 CLI:
```bash
go run ./cmd/chatctl -addr http://127.0.0.1:9099 -token change-me-long-random status
```

## 트러블슈팅
- Windows bind mount + Git: `dubious ownership` -> `git config --global --add safe.directory /workspace`
- Windows 유명 포트 publish 충돌: 기본은 `expose` 권장, 필요 시 `127.0.0.1:대체포트:6379` 사용