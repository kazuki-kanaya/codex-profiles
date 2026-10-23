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

func Add(name string) (err error) {
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
	return RunCodex(profileDir, "login")
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

	return RunCodex(profileDir, args...)
}

func Remove(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}

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
	if err := RunCodex(profileDir, "app-server", "daemon", "stop"); err != nil {
		return fmt.Errorf("stop daemon for profile %q before removing: %w", name, err)
	}

	return os.RemoveAll(profileDir)
}

func List() ([]string, error) {
	baseDir, err := BaseDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(baseDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() || ValidateName(entry.Name()) != nil {
			continue
		}
		names = append(names, entry.Name())
	}
	return names, nil
}

func ValidateRename(oldName, newName string) error {
	for _, name := range []string{oldName, newName} {
		if err := ValidateName(name); err != nil {
			return err
		}
	}
	oldDir, err := ProfileDir(oldName)
	if err != nil {
		return err
	}
	info, err := os.Lstat(oldDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("profile must be a real directory: %s", oldDir)
	}
	archiveDir, err := ArchivedSessionsDir()
	if err != nil {
		return err
	}
	oldArchive := filepath.Join(oldDir, "archived_sessions")
	if _, err := os.Lstat(oldArchive); err == nil {
		target, err := os.Readlink(oldArchive)
		if err != nil {
			return fmt.Errorf("migrate archived sessions from %s to %s before renaming", oldArchive, archiveDir)
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(oldDir, target)
		}
		if filepath.Clean(target) != archiveDir {
			return fmt.Errorf("archived sessions at %s must link to %s before renaming", oldArchive, archiveDir)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	newDir, err := ProfileDir(newName)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(newDir); err == nil {
		return fmt.Errorf("profile %q already exists", newName)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func Rename(oldName, newName string) (err error) {
	if err := ValidateRename(oldName, newName); err != nil {
		return err
	}
	oldDir, err := ProfileDir(oldName)
	if err != nil {
		return err
	}
	newDir, err := ProfileDir(newName)
	if err != nil {
		return err
	}
	if err := RunCodex(oldDir, "app-server", "daemon", "stop"); err != nil {
		return fmt.Errorf("stop daemon for profile %q before renaming: %w", oldName, err)
	}
	if err := os.Mkdir(newDir, 0o700); err != nil {
		return err
	}
	prepared := false
	defer func() {
		if err != nil && !prepared {
			err = errors.Join(err, os.RemoveAll(newDir))
		}
	}()

	for _, name := range []string{"auth.json", "config.toml"} {
		source := filepath.Join(oldDir, name)
		info, err := os.Lstat(source)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("expected a regular file: %s", source)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(newDir, name), data, 0o600); err != nil {
			return err
		}
	}
	if err := linkSessions(newDir); err != nil {
		return err
	}

	prepared = true
	if err := os.RemoveAll(oldDir); err != nil {
		return fmt.Errorf("new profile ready at %s; old profile cleanup failed at %s: %w", newDir, oldDir, err)
	}
	return nil
}
