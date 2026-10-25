package profile

import (
	"fmt"
	"os"
	"path/filepath"
)

func BaseDir() (string, error) {
	if baseDir := os.Getenv("CODEX_PROFILES_HOME"); baseDir != "" {
		return filepath.Abs(baseDir)
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".codex-profiles"), nil
}

func SessionsDir() (string, error) {
	if sessionsDir := os.Getenv("CODEX_PROFILES_SESSIONS_DIR"); sessionsDir != "" {
		return filepath.Abs(sessionsDir)
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".codex", "sessions"), nil
}

func ArchivedSessionsDir() (string, error) {
	if archiveDir := os.Getenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR"); archiveDir != "" {
		return filepath.Abs(archiveDir)
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHomeDir, ".codex", "archived_sessions"), nil
}

func ProfileDir(name string) (string, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, name), nil
}

func linkSessions(profileDir string) error {
	sessionsDir, err := SessionsDir()
	if err != nil {
		return err
	}
	archiveDir, err := ArchivedSessionsDir()
	if err != nil {
		return err
	}
	for name, target := range map[string]string{"sessions": sessionsDir, "archived_sessions": archiveDir} {
		if err := os.MkdirAll(target, 0o700); err != nil {
			return err
		}
		info, err := os.Lstat(target)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return fmt.Errorf("shared %s must be a real directory: %s", name, target)
		}
		if err := os.Symlink(target, filepath.Join(profileDir, name)); err != nil {
			return err
		}
	}
	return nil
}
