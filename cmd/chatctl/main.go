package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	addr := flag.String("addr", "http://127.0.0.1:9099", "admin base url")
	token := flag.String("token", "", "admin token (Bearer)")
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}
	cmd := flag.Arg(0)

	if strings.TrimSpace(*token) == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -token is required")
		os.Exit(2)
	}

	client := &http.Client{Timeout: 6 * time.Second}

	switch cmd {
	case "status":
		doGET(client, *addr+"/admin/status", *token)
	case "stop":
		doPOST(client, *addr+"/admin/stop", *token)
	case "restart":
		doPOST(client, *addr+"/admin/restart", *token)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("usage:")
	fmt.Println("  chatctl -addr http://127.0.0.1:9099 -token <TOKEN> status|stop|restart")
}

func doGET(c *http.Client, url, token string) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.Do(req)
	if err != nil {
		fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	fmt.Print(string(b))
	if res.StatusCode >= 300 {
		os.Exit(1)
	}
}

func doPOST(c *http.Client, url, token string) {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.Do(req)
	if err != nil {
		fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	fmt.Print(string(b))
	if res.StatusCode >= 300 {
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ERROR:", err)
	os.Exit(1)
}
