package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setupProfile(t *testing.T) (oldDir, newDir, sharedDir, stopLog string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake codex requires a POSIX shell")
	}
	root := t.TempDir()
	baseDir := filepath.Join(root, "profiles")
	sharedDir = filepath.Join(root, "shared")
	t.Setenv("CODEX_PROFILES_HOME", baseDir)
	t.Setenv("CODEX_PROFILES_SESSIONS_DIR", filepath.Join(sharedDir, "sessions"))
	t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", filepath.Join(sharedDir, "archived_sessions"))
	oldDir = filepath.Join(baseDir, "old")
	newDir = filepath.Join(baseDir, "new")
	if err := os.MkdirAll(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"auth.json": "test credentials", "config.toml": "cli_auth_credentials_store = \"file\"\n", "state_5.sqlite": "old state"} {
		writeTestFile(t, filepath.Join(oldDir, name), data)
	}
	if err := linkSessions(oldDir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sessions", "archived_sessions"} {
		writeTestFile(t, filepath.Join(sharedDir, name, "keep.jsonl"), name)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.Mkdir(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	stopLog = filepath.Join(root, "stop.log")
	t.Setenv("STOP_LOG", stopLog)
	t.Setenv("STOP_EXIT", "0")
	fake := filepath.Join(binDir, "codex")
	writeTestFile(t, fake, `#!/bin/sh
[ "$*" = "app-server daemon stop" ] || exit 90
[ -f "$CODEX_HOME/auth.json" ] || exit 91
printf '%s\n' "$CODEX_HOME" >> "$STOP_LOG"
exit "$STOP_EXIT"
`)
	if err := os.Chmod(fake, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return oldDir, newDir, sharedDir, stopLog
}

func writeTestFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertTestFile(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("%s = %q, want %q", path, data, want)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be absent, got %v", path, err)
	}
}

func assertSharedHistory(t *testing.T, sharedDir string) {
	t.Helper()
	for _, name := range []string{"sessions", "archived_sessions"} {
		assertTestFile(t, filepath.Join(sharedDir, name, "keep.jsonl"), name)
	}
}

func TestRemovePreservesSharedHistory(t *testing.T) {
	oldDir, _, sharedDir, stopLog := setupProfile(t)
	if err := Remove("old"); err != nil {
		t.Fatal(err)
	}
	assertMissing(t, oldDir)
	assertTestFile(t, stopLog, oldDir+"\n")
	assertSharedHistory(t, sharedDir)
}

func TestRenameCreatesFreshEnvironment(t *testing.T) {
	oldDir, newDir, sharedDir, stopLog := setupProfile(t)
	if err := Rename("old", "new"); err != nil {
		t.Fatal(err)
	}
	assertMissing(t, oldDir)
	assertTestFile(t, stopLog, oldDir+"\n")
	assertTestFile(t, filepath.Join(newDir, "auth.json"), "test credentials")
	assertTestFile(t, filepath.Join(newDir, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
	assertMissing(t, filepath.Join(newDir, "state_5.sqlite"))
	for _, name := range []string{"sessions", "archived_sessions"} {
		target, err := os.Readlink(filepath.Join(newDir, name))
		if err != nil || target != filepath.Join(sharedDir, name) {
			t.Fatalf("%s link = %q, error = %v", name, target, err)
		}
	}
	assertSharedHistory(t, sharedDir)
}

func TestDaemonStopFailurePreservesProfile(t *testing.T) {
	for _, operation := range []string{"remove", "rename"} {
		t.Run(operation, func(t *testing.T) {
			oldDir, newDir, sharedDir, stopLog := setupProfile(t)
			t.Setenv("STOP_EXIT", "1")
			var err error
			if operation == "remove" {
				err = Remove("old")
			} else {
				err = Rename("old", "new")
			}
			if err == nil || !strings.Contains(err.Error(), "stop daemon") {
				t.Fatalf("expected daemon stop error, got %v", err)
			}
			assertTestFile(t, stopLog, oldDir+"\n")
			assertTestFile(t, filepath.Join(oldDir, "auth.json"), "test credentials")
			assertTestFile(t, filepath.Join(oldDir, "state_5.sqlite"), "old state")
			assertMissing(t, newDir)
			assertSharedHistory(t, sharedDir)
		})
	}
}

func TestRenamePreparationFailureRollsBack(t *testing.T) {
	oldDir, newDir, sharedDir, _ := setupProfile(t)
	if err := os.Remove(filepath.Join(oldDir, "config.toml")); err != nil {
		t.Fatal(err)
	}
	if err := Rename("old", "new"); err == nil {
		t.Fatal("expected missing config error")
	}
	assertMissing(t, newDir)
	assertTestFile(t, filepath.Join(oldDir, "auth.json"), "test credentials")
	assertSharedHistory(t, sharedDir)
}

func TestRenameRejectsExistingDestination(t *testing.T) {
	oldDir, newDir, sharedDir, stopLog := setupProfile(t)
	if err := os.Mkdir(newDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(newDir, "keep"), "existing profile")
	if err := Rename("old", "new"); err == nil {
		t.Fatal("expected duplicate profile error")
	}
	assertTestFile(t, filepath.Join(newDir, "keep"), "existing profile")
	assertTestFile(t, filepath.Join(oldDir, "auth.json"), "test credentials")
	assertMissing(t, stopLog)
	assertSharedHistory(t, sharedDir)
}

func TestRemoveRejectsProfileSymlink(t *testing.T) {
	oldDir, newDir, sharedDir, stopLog := setupProfile(t)
	if err := os.Symlink(oldDir, newDir); err != nil {
		t.Fatal(err)
	}
	if err := Remove("new"); err == nil {
		t.Fatal("expected profile symlink rejection")
	}
	assertTestFile(t, filepath.Join(oldDir, "auth.json"), "test credentials")
	assertMissing(t, stopLog)
	assertSharedHistory(t, sharedDir)
}

func TestRenameRejectsPrivateArchive(t *testing.T) {
	oldDir, newDir, sharedDir, stopLog := setupProfile(t)
	archiveDir := filepath.Join(oldDir, "archived_sessions")
	if err := os.Remove(archiveDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(archiveDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(archiveDir, "private.jsonl"), "private history")
	if err := Rename("old", "new"); err == nil {
		t.Fatal("expected private archive rejection")
	}
	assertTestFile(t, filepath.Join(archiveDir, "private.jsonl"), "private history")
	assertMissing(t, newDir)
	assertMissing(t, stopLog)
	assertSharedHistory(t, sharedDir)
}
