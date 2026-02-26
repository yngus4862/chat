package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yngus4862/chat/internal/control"
)

func main() {
	appHTTP := env("APP_HTTP_ADDR", "0.0.0.0:8080")
	appWS := env("APP_WS_ADDR", "0.0.0.0:8081")

	adminAddr := env("ADMIN_HTTP_ADDR", "127.0.0.1:9099")
	adminToken := env("ADMIN_TOKEN", "")

	startedAt := time.Now()

	// ✅ 핵심 변경점 1:
	// 기존의 router.Run(...) 같은 블로킹 실행 대신,
	// http.Server로 감싸서 Shutdown(ctx)를 호출할 수 있게 만든다.
	restSrv := &http.Server{Addr: appHTTP, Handler: buildRestHandler()}
	wsSrv := &http.Server{Addr: appWS, Handler: buildWSHandler()}

	// control event bus
	emitter, sigs := control.New()

	// 앱 전체 컨텍스트(관리 API 종료에도 사용)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	statusFn := func() control.Status {
		return control.BuildStatus(startedAt, appHTTP, appWS, adminAddr)
	}

	// ✅ 핵심 변경점 2: 콘솔 제어(옵션)
	control.StartConsole(ctx, os.Stdin, os.Stdout, emitter, statusFn)

	// ✅ 핵심 변경점 3: 관리 API(토큰 없으면 비활성)
	go func() {
		if adminToken == "" {
			log.Println("[admin] ADMIN_TOKEN empty -> admin server disabled")
			return
		}
		if err := control.StartAdminHTTP(ctx, adminAddr, adminToken, emitter, statusFn); err != nil {
			log.Println("[admin] error:", err)
			emitter.RequestStop()
		}
	}()

	// ✅ 핵심 변경점 4: 서버는 goroutine으로 실행
	go func() {
		log.Println("[app] REST listening:", appHTTP)
		if err := restSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("[app] REST error:", err)
			emitter.RequestStop()
		}
	}()

	go func() {
		log.Println("[app] WS listening:", appWS)
		if err := wsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("[app] WS error:", err)
			emitter.RequestStop()
		}
	}()

	// OS signal (SIGINT/SIGTERM)
	osSig := make(chan os.Signal, 2)
	signal.Notify(osSig, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-osSig:
			log.Println("[ctrl] signal -> stop")
			gracefulStop(restSrv, wsSrv)
			return

		case <-sigs.Stop:
			log.Println("[ctrl] stop requested")
			gracefulStop(restSrv, wsSrv)
			return

		case <-sigs.Restart:
			log.Println("[ctrl] restart requested")
			gracefulStop(restSrv, wsSrv)
			// ✅ 재시작 방식: 현재 프로세스를 동일 바이너리로 교체(re-exec)
			if err := control.ReexecSelf(); err != nil {
				log.Println("[ctrl] reexec failed:", err)
				return
			}
		}
	}
}

func gracefulStop(restSrv, wsSrv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = wsSrv.Shutdown(ctx)
	_ = restSrv.Shutdown(ctx)
}

func env(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// TODO: 여기만 “당신이 이미 만든 라우터/핸들러”로 교체
func buildRestHandler() http.Handler { return http.NewServeMux() }
func buildWSHandler() http.Handler   { return http.NewServeMux() }
