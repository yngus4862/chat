package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yngus4862/chat/internal/config"
	"github.com/yngus4862/chat/internal/httpapi"
	"github.com/yngus4862/chat/internal/realtime"
	"github.com/yngus4862/chat/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer st.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := st.EnsureSchema(ctx); err != nil {
		log.Fatalf("ensure schema failed: %v", err)
	}

	hub := realtime.NewHub()

	restHandler := httpapi.NewRouter(st, hub)
	restSrv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           restHandler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", realtime.WSHandler(hub, st))
	wsSrv := &http.Server{
		Addr:              ":" + cfg.WsPort,
		Handler:           wsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("REST listening on :%s", cfg.AppPort)
		if err := restSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("REST server failed: %v", err)
		}
	}()

	go func() {
		log.Printf("WS listening on :%s", cfg.WsPort)
		if err := wsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("WS server failed: %v", err)
		}
	}()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	_ = restSrv.Shutdown(shutdownCtx)
	_ = wsSrv.Shutdown(shutdownCtx)

	log.Println("shutdown complete")
}