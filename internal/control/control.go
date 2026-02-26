package control

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Status struct {
	PID       int       `json:"pid"`
	UptimeSec int64     `json:"uptimeSec"`
	StartedAt time.Time `json:"startedAt"`
	AppHTTP   string    `json:"appHttp,omitempty"`
	AppWS     string    `json:"appWs,omitempty"`
	Admin     string    `json:"admin,omitempty"`
	GoVersion string    `json:"goVersion"`
	OS        string    `json:"os"`
	Arch      string    `json:"arch"`
}

type Emitter struct {
	stop    chan struct{}
	restart chan struct{}
}

type Signals struct {
	Stop    <-chan struct{}
	Restart <-chan struct{}
}

func New() (*Emitter, Signals) {
	e := &Emitter{
		stop:    make(chan struct{}, 1),
		restart: make(chan struct{}, 1),
	}
	return e, Signals{Stop: e.stop, Restart: e.restart}
}

// 중복 요청은 “대기 중 1개”만 유지(버퍼 1). 처리되면 다시 받을 수 있음.
func (e *Emitter) RequestStop() {
	select {
	case e.stop <- struct{}{}:
	default:
	}
}

func (e *Emitter) RequestRestart() {
	select {
	case e.restart <- struct{}{}:
	default:
	}
}

func BuildStatus(startedAt time.Time, appHTTP, appWS, admin string) Status {
	uptime := int64(time.Since(startedAt).Seconds())
	return Status{
		PID:       os.Getpid(),
		UptimeSec: uptime,
		StartedAt: startedAt,
		AppHTTP:   appHTTP,
		AppWS:     appWS,
		Admin:     admin,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// Linux 컨테이너 기준: 현재 프로세스를 동일 바이너리로 교체(re-exec)
func ReexecSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := append([]string{exe}, os.Args[1:]...)
	env := os.Environ()
	return syscall.Exec(exe, args, env)
}

func StartAdminHTTP(ctx context.Context, addr string, token string, e *Emitter, statusFn func() Status) error {
	if strings.TrimSpace(addr) == "" {
		return errors.New("admin addr is empty")
	}
	if strings.TrimSpace(token) == "" {
		return errors.New("admin token is empty (refuse to start admin server)")
	}

	mux := http.NewServeMux()

	auth := func(w http.ResponseWriter, r *http.Request) bool {
		ah := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if !strings.HasPrefix(ah, prefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return false
		}
		got := strings.TrimSpace(strings.TrimPrefix(ah, prefix))
		if got != token {
			http.Error(w, "forbidden", http.StatusForbidden)
			return false
		}
		return true
	}

	mux.HandleFunc("/admin/status", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		writeJSON(w, statusFn())
	})

	mux.HandleFunc("/admin/stop", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		e.RequestStop()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("stopping\n"))
	})

	mux.HandleFunc("/admin/restart", func(w http.ResponseWriter, r *http.Request) {
		if !auth(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		e.RequestRestart()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("restarting\n"))
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func StartConsole(ctx context.Context, in io.Reader, out io.Writer, e *Emitter, statusFn func() Status) {
	// air 사용 시 stdin이 air에 먹힐 수 있으니 “옵션”으로 둠.
	go func() {
		sc := bufio.NewScanner(in)
		fmt.Fprintln(out, "control console: type 'help'")
		for sc.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			cmd := strings.TrimSpace(sc.Text())
			switch cmd {
			case "":
				continue
			case "help":
				fmt.Fprintln(out, "commands: status | stop | restart | help")
			case "status":
				b, _ := json.MarshalIndent(statusFn(), "", "  ")
				fmt.Fprintln(out, string(b))
			case "stop":
				e.RequestStop()
				fmt.Fprintln(out, "stop requested")
			case "restart":
				e.RequestRestart()
				fmt.Fprintln(out, "restart requested")
			default:
				fmt.Fprintln(out, "unknown command:", cmd)
			}
		}
	}()
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
