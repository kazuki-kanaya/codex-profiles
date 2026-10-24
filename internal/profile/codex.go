package profile

import (
	"context"
	"os"
	"os/exec"
)

func RunCodex(profileDir string, args ...string) error {
	return runCodex(context.Background(), profileDir, args...)
}

func runCodex(ctx context.Context, profileDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Env = append(os.Environ(), "CODEX_HOME="+profileDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
