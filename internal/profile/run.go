package profile

import (
	"context"
	"fmt"
	"os"

	"github.com/kazuki-kanaya/codex-profiles/internal/codex"
)

func Run(name string, args ...string) error {
	profileDir, err := ProfileDir(name)
	if err != nil {
		return err
	}

	info, err := os.Lstat(profileDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("profile must be a real directory: %s", profileDir)
	}

	return codex.Run(context.Background(), profileDir, args...)
}
