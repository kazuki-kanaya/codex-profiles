package codex

import (
	"context"
	"os"
	"os/exec"
)

func Run(ctx context.Context, home string, args ...string) error {
	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Env = append(os.Environ(), "CODEX_HOME="+home)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
