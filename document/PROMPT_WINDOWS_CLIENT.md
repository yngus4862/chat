# 구조화된 프롬프트: Windows 클라이언트(.NET MAUI) 개발 (향후 Android/iOS 확장 포함)

## 0) 역할(Role)

너는 **시니어 클라이언트 엔지니어 + 모바일/데스크톱 아키텍트 + OIDC 인증 설계자**다.
목표는 “지금은 Windows부터, 하지만 곧 Android/iOS도 같은 코드베이스로” 갈 수 있게 **처음부터 크로스플랫폼 구조**로 구현하는 것이다.

## 1) 목표(Goal)

다음 레포의 서버와 연동되는 **Windows 클라이언트 앱**을 개발한다:

* 소스: `https://github.com/yngus4862/chat`

필수 요구:

1. 서버 이벤트(실시간)를 받아 **Windows 시스템 알림(토스트)** 으로 사용자에게 즉시 알린다.
2. **크로스플랫폼 언어/프레임워크**를 사용한다(향후 Android/iOS 개발 고려).
3. 서버는 REST + WebSocket 기반이며, 최소한 MVP 엔드포인트를 모두 동작시킨다.

## 2) 기술 선택(Decision)

### 2-1. 1안(권장): .NET MAUI (C#)

* 장점: Windows/Android/iOS를 하나의 C# 코드베이스로 커버 가능. ([Microsoft Learn][2])
* Windows 토스트: Windows App SDK 알림(Toast/App Notifications) 방식 적용. ([Microsoft Learn][3])

### 2-2. 2안(대안): Flutter (Dart)

* 장점: UI 생산성/일관성 매우 좋음
* 단점: Windows 알림, OIDC, WS 등 “네이티브 브릿지” 구성에 손이 더 갈 수 있음

> 이번 구현은 1안(.NET MAUI)으로 진행하되, 설계는 “플랫폼별 알림/푸시 추상화”로 2안도 수용 가능하게 만든다.

## 3) 입력(Input) — 여기 값이 없으면 TODO로 남겨라

### 3-1. 서버 접속 정보

* DEV REST BaseUrl: `http://localhost:8080` (README 기준) ([GitHub][1])
* DEV WS Url: `ws://localhost:8081/ws?roomId={roomId}` (README 기준) ([GitHub][1])
* PROD(리버스프록시 사용 시):

  * REST: `https://<host>/api`
  * WS: `wss://<host>/realtime` (또는 `/ws`)
  * (실제 경로는 레포/설정에 맞춰 자동 탐지 또는 설정 파일로 주입)

### 3-2. 인증(OIDC / Keycloak)

* Issuer URL: `https://<keycloak-host>/realms/<realm>`
* ClientId: `<client-id>`
* Redirect URI(앱 스킴): `mychat://callback` (플랫폼별 등록 필요)
* Flow: Authorization Code + PKCE 권장. ([Keycloak][4])

## 4) 기능 범위(Scope)

### 4-1. MVP 기능

* 로그인/로그아웃(OIDC)
* 채팅방 목록 조회, 방 생성
* 메시지 목록 조회(최근 N개), 추가 로딩(페이징)
* 메시지 전송
* WebSocket 실시간 수신(해당 roomId)
* **새 메시지 수신 시 Windows 토스트 알림 표시**
* 최소 로컬 캐시(선택): 마지막 동기화 시각, 최근 방/메시지 일부

### 4-2. 확장 고려(설계에 훅만 심기)

* 멀티룸 구독(서버가 현재 `roomId` 기반이면, “백그라운드에서 관심 방만 WS 연결” 전략)
* 오프라인 푸시:

  * Windows: WNS(추후), 현재는 “온라인 실시간 + 토스트” 우선
  * Android: FCM / iOS: APNs (PushToken 등록 API가 생기면 연결)

## 5) 아키텍처 요구사항(중요)

### 5-1. 계층 분리

* Presentation(UI): MVVM
* Application: 유스케이스(JoinRoom, SendMessage, SyncMessages…)
* Infrastructure: REST/WS/OIDC/Storage/Notifications 구현체

### 5-2. “알림” 추상화(핵심)

* `INotificationService`

  * `ShowMessageToast(roomName, sender, previewText, deepLink)`
* Windows 구현: Windows App SDK 알림(Toast/App Notifications) ([Microsoft Learn][3])
* Android/iOS 구현은 추후 추가 가능하도록 인터페이스만 고정

### 5-3. 재연결/신뢰성

* WebSocket 끊김 감지 + **지수 백오프 재연결**
* 중복 수신 대비(서버가 at-least-once면): `message_id` 또는 `clientMsgId` 기준 dedupe
* 백그라운드/최소화 상태에서 알림 스팸 방지(예: 같은 방 연속 N개는 “요약 알림”)

## 6) 구현 단계(Plan)

### 6-1. 서버 계약(Contract) 먼저 확인

1. README에 명시된 REST/WS 엔드포인트를 기준으로 “요청/응답 스키마”를 추출한다. ([GitHub][1])
2. 서버 코드에서 실제 JSON 필드명을 확인하고, `ChatApiClient` DTO에 반영한다.
3. 스키마가 불명확하면 “서버에서 실제 응답 샘플(JSON)”을 출력하도록 요청(또는 smoketest 활용).

### 6-2. 프로젝트 스캐폴딩

* 솔루션 예시:

  * `Chat.Client` (MAUI UI)
  * `Chat.Client.Core` (UseCase/Domain/Interfaces)
  * `Chat.Client.Infrastructure` (REST/WS/OIDC/SQLite/Notification 구현)
  * `Chat.Client.Tests` (unit tests)

### 6-3. OIDC 로그인

* 시스템 브라우저 기반 로그인 + PKCE
* 토큰 저장:

  * Windows: DPAPI 보호 또는 OS 안전 저장소
  * MAUI 공통: SecureStorage 사용(가능 범위 내)

### 6-4. REST 연동

* `GET /v1/rooms`
* `POST /v1/rooms`
* `GET /v1/rooms/{roomId}/messages?limit=50`
* `POST /v1/rooms/{roomId}/messages`

### 6-5. WebSocket + 토스트

* `ws://.../ws?roomId=...` 연결 ([GitHub][1])
* 메시지 수신 이벤트 처리:

  * 앱 포그라운드 + 해당 방 열람 중이면 “UI만 업데이트”
  * 그 외에는 `INotificationService.ShowMessageToast(...)` 호출
* Windows 토스트 구현은 Windows App SDK 가이드 패턴 준수(등록/해제 포함). ([Microsoft Learn][3])

## 7) 산출물(Output) — 너는 아래를 **한 번에** 내놔야 한다

1. 최종 아키텍처 다이어그램(텍스트) + 선택한 기술스택 요약
2. 폴더 트리 + 핵심 파일 목록
3. 실행 가능한 코드:

   * MAUI 앱 엔트리, DI 설정
   * OIDC 로그인 플로우(샘플)
   * REST 클라이언트
   * WebSocket 클라이언트(재연결 포함)
   * Windows 토스트 알림 구현(작동 예시 포함)
   * 최소 UI(방 목록 / 메시지 목록 / 전송)
4. 로컬 실행 방법:

   * 서버 실행(DevContainer 기준)
   * 클라이언트 실행(Visual Studio / CLI)
   * 통합 테스트 시나리오(“방 만들고 → 메시지 보내고 → 다른 창에서 토스트 뜨는지”)

## 8) 품질 기준(Quality Bar)

* “돌아간다”가 아니라, **재연결/예외/로그/알림 스팸 방지**까지 기본 탑재
* 하드코딩 금지: BaseUrl, WS Url, OIDC 설정은 `appsettings.json` + 환경변수로 주입
* 코드에는 “Android/iOS 확장 지점”을 명확히 주석/인터페이스로 남긴다

[1]: https://raw.githubusercontent.com/yngus4862/chat/main/README.md "raw.githubusercontent.com"
[2]: https://learn.microsoft.com/en-us/dotnet/maui/supported-platforms?view=net-maui-10.0&utm_source=chatgpt.com "Supported platforms for .NET MAUI apps"
[3]: https://learn.microsoft.com/en-us/windows/apps/develop/notifications/app-notifications/app-notifications-quickstart?utm_source=chatgpt.com "Quickstart App notifications in the Windows App SDK"
[4]: https://www.keycloak.org/securing-apps/oidc-layers?utm_source=chatgpt.com "Securing applications and services with OpenID Connect"
