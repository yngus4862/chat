# Go Chat Server Skeleton (Dev Container)

## Why this layout?
Go는 “딱 하나의 표준 레이아웃”을 강제하진 않지만, 공식 문서(go.dev)가 서버 프로젝트에 대해
`go.mod`는 루트, 실행 바이너리는 `cmd/`, 서버 로직은 `internal/`에 두는 구성을 권장합니다.  
(내부 패키지는 외부 모듈에서 import가 금지되어 리팩터링 자유도가 큽니다.)

## Tree
workspace/
  cmd/chatd/main.go
  internal/{config,model,store,httpapi,realtime}
  migrations/

## Endpoints
REST: http://localhost:8080
- GET /healthz
- GET /readyz
- POST /v1/rooms
- GET /v1/rooms
- POST /v1/rooms/{roomId}/messages
- GET /v1/rooms/{roomId}/messages?limit=50

WebSocket: ws://localhost:8081/ws?roomId=1

## Run
- VS Code: Reopen in Container
- CLI: cd workspace/.devcontainer && docker compose up --build

## Note (Conflict)
PROJECT_STATE.md의 확정 스택은 .NET/SignalR 입니다.
본 Go 스켈레톤은 “요청에 의해” 생성된 대안이며,
현재는 구조/개발환경 재현성 확인용으로만 유지하는 것을 권장합니다.