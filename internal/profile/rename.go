package profile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
