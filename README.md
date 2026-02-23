# 사내 메신저 인프라 스택 (앱은 Host 실행)

## 핵심
- 외부 노출: **443만**
- nginx(TLS 종료) -> Keycloak(/auth), Host 앱(/api, /realtime)
- 내부 서비스(Postgres/Redis/MinIO/Elastic/Observability)는 Docker 네트워크로만

## 1) 준비
1) env 생성
- `.devcontainer/.env.example` -> `.devcontainer/.env` 복사 후 비밀번호 변경

2) 데이터 디렉터리
- 권장: `workspace/.data/*`

3) TLS
- 개발용(self-signed) 생성(선택)
  - WSL/Linux: `bash .devcontainer/scripts/generate-selfsigned.sh`

## 2) 실행
루트에서 실행(권장):
```bash
docker compose -f .devcontainer/compose.yaml --env-file .devcontainer/.env up -d
