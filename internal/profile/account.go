package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func ValidateName(name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("invalid profile name %q: start with a lowercase letter (a-z). The remaining characters can be a-z, 0-9, '-' or '_' (example: work-2)", name)
	}
	return nil
}

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

func Add(name string) error {
	sessionsDir, err := SessionsDir()
	if err != nil {
		return err
	}
	info, err := os.Lstat(sessionsDir)
	if err != nil {
		return fmt.Errorf("cannot access shared sessions: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("shared sessions must be a real directory: %s", sessionsDir)
	}

	baseDir, err := BaseDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		return err
	}

	profileDir := filepath.Join(baseDir, name)
	if err := os.Mkdir(profileDir, 0o700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("profile %q already exists", name)
		}
		return err
	}

	if err := os.Symlink(
		sessionsDir,
		filepath.Join(profileDir, "sessions"),
	); err != nil {
		return errors.Join(err, os.Remove(profileDir))
	}

	configPath := filepath.Join(profileDir, "config.toml")
	config := []byte("cli_auth_credentials_store = \"file\"\n")
	if err = os.WriteFile(configPath, config, 0o600); err != nil {
		return errors.Join(err, os.RemoveAll(profileDir))
	}
	if err := RunCodex(profileDir, "login"); err != nil {
		return errors.Join(err, os.RemoveAll(profileDir))
	}
	return nil
}

func Run(name string, args ...string) error {
	baseDir, err := BaseDir()
	if err != nil {
		return err
	}

	profileDir := filepath.Join(baseDir, name)
	info, err := os.Lstat(profileDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("profile must be a real directory: %s", profileDir)
	}

	return RunCodex(profileDir, args...)
}
