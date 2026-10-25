package profile

import (
	"fmt"
	"os"
)

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
