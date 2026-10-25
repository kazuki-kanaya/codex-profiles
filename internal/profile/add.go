package profile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kazuki-kanaya/codex-profiles/internal/codex"
)

func Add(name string) (err error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	baseDir, err := BaseDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return err
	}

	profileDir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	if err := os.Mkdir(profileDir, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("profile %q already exists", name)
		}
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(profileDir))
		}
	}()

	if err := linkSessions(profileDir); err != nil {
		return err
	}

	configPath := filepath.Join(profileDir, "config.toml")
	config := []byte("cli_auth_credentials_store = \"file\"\n")
	if err = os.WriteFile(configPath, config, 0o600); err != nil {
		return err
	}
	if err := codex.Run(ctx, profileDir, "login"); err != nil {
		return err
	}
	return ctx.Err()
}
