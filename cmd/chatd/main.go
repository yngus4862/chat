package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yngus4862/chat/internal/api"
	"github.com/yngus4862/chat/internal/config"
	"github.com/yngus4862/chat/internal/control"
	"github.com/yngus4862/chat/internal/db"
	"github.com/yngus4862/chat/internal/health"
	"github.com/yngus4862/chat/internal/store"
	"github.com/yngus4862/chat/internal/ws"
)

func main() {
	cfg := config.Load()
	startedAt := time.Now()

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// DB
	dbConn, err := db.Connect(rootCtx, cfg.PostgresURL())
	if err != nil {
		log.Fatal("db connect failed: ", err)
	}
	defer dbConn.Close()

	st := store.New(dbConn.Pool)

	// Redis pubsub for WS (optional but recommended)
	var ps *ws.RedisPubSub
	ps = ws.NewRedisPubSub(cfg.RedisAddr())
	defer func() { _ = ps.Close() }()

	hub := ws.NewHub(st, ps)

	// Readiness
	readyFn := func() health.Result {
		return health.Ready(rootCtx, st, ps)
	}

	h := &api.Handlers{Store: st, Hub: hub}
	router := api.NewRouter(api.Deps{Handlers: h, ReadyFn: readyFn})

	restSrv := &http.Server{Addr: cfg.AppHTTPAddr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", hub.ServeWS)
	wsSrv := &http.Server{Addr: cfg.AppWSAddr, Handler: wsMux, ReadHeaderTimeout: 5 * time.Second}

	// Control
	emitter, sigs := control.New()
	statusFn := func() control.Status {
		return control.BuildStatus(startedAt, cfg.AppHTTPAddr, cfg.AppWSAddr, cfg.AdminHTTPAddr)
	}
	control.StartConsole(rootCtx, os.Stdin, os.Stdout, emitter, statusFn)

	go func() {
		if cfg.AdminToken == "" {
			log.Println("[admin] ADMIN_TOKEN empty -> admin server disabled")
			return
		}
		if err := control.StartAdminHTTP(rootCtx, cfg.AdminHTTPAddr, cfg.AdminToken, emitter, statusFn); err != nil {
			log.Println("[admin] error:", err)
			emitter.RequestStop()
		}
	}()

	// Start servers
	go func() {
		log.Println("[app] REST listening:", cfg.AppHTTPAddr)
		if err := restSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("[app] REST error:", err)
			emitter.RequestStop()
		}
	}()
	go func() {
		log.Println("[app] WS listening:", cfg.AppWSAddr)
		if err := wsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("[app] WS error:", err)
			emitter.RequestStop()
		}
	}()

	// OS signals
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
