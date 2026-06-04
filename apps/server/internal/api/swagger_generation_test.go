package api

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGeneratedSwaggerDocsAreFresh(t *testing.T) {
	if testing.Short() {
		t.Skip("skip generated Swagger freshness check in short mode")
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	repoRoot := filepath.Clean(filepath.Join(cwd, "../../../.."))
	scriptPath := filepath.Join(repoRoot, "scripts/generate-server-swagger.sh")
	command := exec.Command(scriptPath, "--check")
	command.Dir = repoRoot
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.24.4")

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("generated Swagger docs are stale or cannot be regenerated: %v\n%s", err, string(output))
	}
}
