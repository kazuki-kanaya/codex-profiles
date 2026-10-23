package profile

import (
	"os"
	"os/exec"
)

func RunCodex(profileDir string, args ...string) error {
	cmd := exec.Command("codex", args...)
	cmd.Env = append(os.Environ(), "CODEX_HOME="+profileDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
