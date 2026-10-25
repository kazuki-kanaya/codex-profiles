package profile

import (
	"errors"
	"os"
)

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
