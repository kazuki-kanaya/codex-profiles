package profile

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

func TestRename(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("daemon helper requires a POSIX shell")
	}
	tests := []struct {
		name         string
		sourceFile   string
		sourceKind   string
		blocked      string
		destination  bool
		daemonFails  bool
		codexMissing bool
		wantErr      bool
	}{
		{name: "rename profile"},
		{name: "daemon stop fails", daemonFails: true, wantErr: true},
		{name: "codex executable missing", codexMissing: true, wantErr: true},
		{name: "destination exists", destination: true, wantErr: true},
		{name: "auth missing", sourceFile: "auth.json", sourceKind: "missing", wantErr: true},
		{name: "config missing", sourceFile: "config.toml", sourceKind: "missing", wantErr: true},
		{name: "auth is a directory", sourceFile: "auth.json", sourceKind: "directory", wantErr: true},
		{name: "config is a directory", sourceFile: "config.toml", sourceKind: "directory", wantErr: true},
		{name: "auth is a symlink", sourceFile: "auth.json", sourceKind: "symlink", wantErr: true},
		{name: "config is a symlink", sourceFile: "config.toml", sourceKind: "symlink", wantErr: true},
		{name: "sessions link preparation fails", blocked: "sessions", wantErr: true},
		{name: "archive link preparation fails", blocked: "archived_sessions", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			baseDir := filepath.Join(root, "profiles")
			oldDir := filepath.Join(baseDir, "main")
			newDir := filepath.Join(baseDir, "work")
			bin := filepath.Join(root, "bin")
			logPath := filepath.Join(t.TempDir(), "daemon.log")
			links := map[string]string{
				"sessions":          filepath.Join(root, "sessions"),
				"archived_sessions": filepath.Join(root, "archived_sessions"),
			}
			for _, dir := range []string{oldDir, bin, filepath.Join(baseDir, "other"), links["sessions"], links["archived_sessions"]} {
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			files := map[string]string{
				filepath.Join(oldDir, "auth.json"):                          "test credentials",
				filepath.Join(oldDir, "config.toml"):                        "cli_auth_credentials_store = \"file\"\n",
				filepath.Join(oldDir, "state_5.sqlite"):                     "old index",
				filepath.Join(oldDir, "daemon-state"):                       "old daemon state",
				filepath.Join(baseDir, "other", "auth.json"):                "other credentials",
				filepath.Join(links["sessions"], "keep.jsonl"):              "shared session",
				filepath.Join(links["archived_sessions"], "keep.jsonl.zst"): "shared archive",
			}
			for path, content := range files {
				if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tt.sourceFile != "" {
				path := filepath.Join(oldDir, tt.sourceFile)
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				switch tt.sourceKind {
				case "directory":
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				case "symlink":
					if err := os.Symlink(filepath.Join(baseDir, "other", "auth.json"), path); err != nil {
						t.Fatal(err)
					}
				}
			}
			if tt.destination {
				if err := os.Mkdir(newDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(newDir, "auth.json"), []byte("existing credentials"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if tt.blocked != "" {
				path := filepath.Join(root, "blocked")
				if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
				links[tt.blocked] = path
			}
			for name, target := range links {
				if err := os.Symlink(target, filepath.Join(oldDir, name)); err != nil {
					t.Fatal(err)
				}
			}
			helper := "#!/bin/sh\nprintf '%s\\n' \"$CODEX_HOME\" \"$@\" > \"$CODEX_PROFILES_TEST_LOG\"\n[ -d \"$CODEX_HOME\" ] || exit 99\nexit \"$CODEX_PROFILES_TEST_EXIT\"\n"
			if !tt.codexMissing {
				if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(helper), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("HOME", root)
			t.Setenv("CODEX_HOME", filepath.Join(root, "parent-home"))
			t.Setenv("CODEX_PROFILES_HOME", baseDir)
			t.Setenv("CODEX_PROFILES_SESSIONS_DIR", links["sessions"])
			t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", links["archived_sessions"])
			t.Setenv("PATH", bin)
			t.Setenv("CODEX_PROFILES_TEST_LOG", logPath)
			t.Setenv("CODEX_PROFILES_TEST_EXIT", "0")
			if tt.daemonFails {
				t.Setenv("CODEX_PROFILES_TEST_EXIT", "17")
			}
			before := renameTestState(t, root)

			err := Rename("main", "work")

			if (err != nil) != tt.wantErr {
				t.Errorf("Rename() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.daemonFails {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 17 {
					t.Errorf("expected daemon exit code 17, got %v", err)
				}
			}
			log, logErr := os.ReadFile(logPath)
			if tt.destination || tt.codexMissing {
				if !errors.Is(logErr, os.ErrNotExist) {
					t.Errorf("expected no daemon invocation, got %q, error = %v", log, logErr)
				}
			} else {
				want := strings.Join([]string{oldDir, "app-server", "daemon", "stop", ""}, "\n")
				if logErr != nil || string(log) != want {
					t.Errorf("daemon invocation = %q, error = %v; want %q", log, logErr, want)
				}
			}
			after := renameTestState(t, root)
			if tt.wantErr {
				if !maps.Equal(before, after) {
					t.Errorf("failed rename changed files: before = %v, after = %v", before, after)
				}
				return
			}
			if _, err := os.Lstat(oldDir); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("expected old profile to be absent, got %v", err)
			}
			entries, err := os.ReadDir(newDir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 4 {
				t.Errorf("new profile has %d entries, want only auth, config, and two shared links", len(entries))
			}
			for _, name := range []string{"auth.json", "config.toml"} {
				path := filepath.Join(newDir, name)
				if got, err := os.ReadFile(path); err != nil || string(got) != files[filepath.Join(oldDir, name)] {
					t.Errorf("copied %s = %q, error = %v", name, got, err)
				}
			}
			for name, want := range links {
				if got, err := os.Readlink(filepath.Join(newDir, name)); err != nil || got != want {
					t.Errorf("%s link = %q, error = %v; want %q", name, got, err, want)
				}
			}
			for _, path := range []string{newDir, filepath.Join(newDir, "auth.json"), filepath.Join(newDir, "config.toml")} {
				info, err := os.Lstat(path)
				if err != nil {
					t.Fatal(err)
				}
				if path == newDir && !info.IsDir() || path != newDir && !info.Mode().IsRegular() {
					t.Errorf("unexpected file type at %s: %s", path, info.Mode())
				}
				if info.Mode().Perm()&0o077 != 0 {
					t.Errorf("%s is accessible by other users: %o", path, info.Mode().Perm())
				}
			}
			for path := range before {
				if path == oldDir || strings.HasPrefix(path, oldDir+string(os.PathSeparator)) {
					delete(before, path)
				}
			}
			for path := range after {
				if path == newDir || strings.HasPrefix(path, newDir+string(os.PathSeparator)) {
					delete(after, path)
				}
			}
			if !maps.Equal(before, after) {
				t.Errorf("rename changed unrelated files: before = %v, after = %v", before, after)
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
