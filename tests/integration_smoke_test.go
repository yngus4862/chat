//go:build integration

package tests

import (
	"os/exec"
	"testing"
)

func TestSmoke(t *testing.T) {
	cmd := exec.Command("go", "run", "./cmd/smoketest")
	cmd.Env = append(cmd.Environ(),
		"BASE_URL=http://localhost:8080",
		"WS_URL=ws://localhost:8081/ws",
		"SMOKE_TIMEOUT_SEC=20",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("smoke failed: %v\n%s", err, string(out))
	}
	t.Log(string(out))
}
