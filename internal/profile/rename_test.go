package profile

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRename(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		newName string
		oldKind string
		newKind string
		archive string
		wantErr bool
	}{
		{name: "absolute archive link", oldName: "main", newName: "work"},
		{name: "relative archive link", oldName: "main", newName: "work", archive: "relative"},
		{name: "archive missing", oldName: "main", newName: "work", archive: "missing"},
		{name: "invalid old name", oldName: "../outside", newName: "work", wantErr: true},
		{name: "invalid new name", oldName: "main", newName: "../new-outside", wantErr: true},
		{name: "empty old name", newName: "work", wantErr: true},
		{name: "empty new name", oldName: "main", wantErr: true},
		{name: "same name", oldName: "main", newName: "main", wantErr: true},
		{name: "source missing", oldName: "main", newName: "work", oldKind: "missing", wantErr: true},
		{name: "source is a file", oldName: "main", newName: "work", oldKind: "file", wantErr: true},
		{name: "source is a symlink", oldName: "main", newName: "work", oldKind: "symlink", wantErr: true},
		{name: "source is a dangling symlink", oldName: "main", newName: "work", oldKind: "dangling symlink", wantErr: true},
		{name: "destination is a directory", oldName: "main", newName: "work", newKind: "directory", wantErr: true},
		{name: "destination is a file", oldName: "main", newName: "work", newKind: "file", wantErr: true},
		{name: "destination is a symlink", oldName: "main", newName: "work", newKind: "symlink", wantErr: true},
		{name: "destination is a dangling symlink", oldName: "main", newName: "work", newKind: "dangling symlink", wantErr: true},
		{name: "archive links elsewhere", oldName: "main", newName: "work", archive: "other", wantErr: true},
		{name: "archive is a directory", oldName: "main", newName: "work", archive: "directory", wantErr: true},
		{name: "archive is a file", oldName: "main", newName: "work", archive: "file", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			baseDir := filepath.Join(root, "profiles")
			oldDir := filepath.Join(baseDir, "main")
			newDir := filepath.Join(baseDir, "work")
			archiveDir := filepath.Join(root, "shared-archive")
			t.Setenv("CODEX_PROFILES_HOME", baseDir)
			t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", archiveDir)
			for _, dir := range []string{baseDir, archiveDir, filepath.Join(root, "outside")} {
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			sharedFile := filepath.Join(archiveDir, "keep.jsonl.zst")
			if err := os.WriteFile(sharedFile, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			oldKind := tt.oldKind
			if oldKind == "" {
				oldKind = "directory"
			}
			for path, kind := range map[string]string{oldDir: oldKind, newDir: tt.newKind} {
				switch kind {
				case "directory":
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(path, "auth.json"), []byte("keep"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "file":
					if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "symlink", "dangling symlink":
					target := archiveDir
					if kind == "dangling symlink" {
						target = filepath.Join(root, "missing")
					}
					if err := os.Symlink(target, path); err != nil {
						t.Fatal(err)
					}
				}
			}
			if tt.oldKind == "" {
				path := filepath.Join(oldDir, "archived_sessions")
				switch tt.archive {
				case "missing":
				case "directory":
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				case "file":
					if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
						t.Fatal(err)
					}
				default:
					target := archiveDir
					if tt.archive == "relative" {
						var err error
						target, err = filepath.Rel(oldDir, archiveDir)
						if err != nil {
							t.Fatal(err)
						}
					} else if tt.archive == "other" {
						target = root
					}
					if err := os.Symlink(target, path); err != nil {
						t.Fatal(err)
					}
				}
			}

			before := renameTestState(t, root)

			err := ValidateRename(tt.oldName, tt.newName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRename(%q, %q) error = %v, wantErr = %v", tt.oldName, tt.newName, err, tt.wantErr)
			}
			if after := renameTestState(t, root); !maps.Equal(before, after) {
				t.Errorf("validation changed files: before = %v, after = %v", before, after)
			}
		})
	}
}

func renameTestState(t *testing.T, root string) map[string]string {
	t.Helper()
	state := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		var content string
		switch {
		case info.Mode().IsRegular():
			data, readErr := os.ReadFile(path)
			content, err = string(data), readErr
		case info.Mode()&os.ModeSymlink != 0:
			content, err = os.Readlink(path)
		}
		state[path] = info.Mode().String() + ":" + content
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return state
}
