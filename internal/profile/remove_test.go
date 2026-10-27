package profile

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRemove(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("daemon helper requires a POSIX shell")
	}
	tests := []struct {
		name         string
		profileName  string
		kind         string
		daemonFails  bool
		codexMissing bool
		wantErr      bool
		wantDaemon   bool
	}{
		{name: "remove profile", profileName: "main", wantDaemon: true},
		{name: "daemon stop fails", profileName: "main", daemonFails: true, wantErr: true, wantDaemon: true},
		{name: "codex executable missing", profileName: "main", codexMissing: true, wantErr: true},
		{name: "profile missing", profileName: "main", kind: "missing", wantErr: true},
		{name: "profile is a file", profileName: "main", kind: "file", wantErr: true},
		{name: "profile is a symlink", profileName: "main", kind: "symlink", wantErr: true},
		{name: "profile is a dangling symlink", profileName: "main", kind: "dangling symlink", wantErr: true},
		{name: "empty name", wantErr: true},
		{name: "parent directory", profileName: "..", wantErr: true},
		{name: "path outside base directory", profileName: "../outside", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			baseDir := filepath.Join(root, "profiles")
			profileDir := filepath.Join(baseDir, "main")
			sessionsDir := filepath.Join(root, "sessions")
			archiveDir := filepath.Join(root, "archived_sessions")
			outsideDir := filepath.Join(root, "outside")
			bin := filepath.Join(root, "bin")
			logPath := filepath.Join(root, "daemon.log")
			preserved := []string{
				filepath.Join(baseDir, "other", "auth.json"),
				filepath.Join(sessionsDir, "keep.jsonl"),
				filepath.Join(archiveDir, "keep.jsonl.zst"),
				filepath.Join(outsideDir, "auth.json"),
			}
			for _, path := range preserved {
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(bin, 0o700); err != nil {
				t.Fatal(err)
			}
			helper := "#!/bin/sh\nprintf '%s\\n' \"$CODEX_HOME\" \"$@\" > \"$CODEX_PROFILES_TEST_LOG\"\n[ -f \"$CODEX_HOME/auth.json\" ] || exit 99\nexit \"$CODEX_PROFILES_TEST_EXIT\"\n"
			if !tt.codexMissing {
				if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(helper), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("HOME", root)
			t.Setenv("CODEX_HOME", filepath.Join(root, "parent-home"))
			t.Setenv("CODEX_PROFILES_HOME", baseDir)
			t.Setenv("PATH", bin)
			t.Setenv("CODEX_PROFILES_TEST_LOG", logPath)
			t.Setenv("CODEX_PROFILES_TEST_EXIT", "0")
			if tt.daemonFails {
				t.Setenv("CODEX_PROFILES_TEST_EXIT", "17")
			}
			switch tt.kind {
			case "missing":
			case "file":
				if err := os.WriteFile(profileDir, []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "symlink", "dangling symlink":
				target := outsideDir
				if tt.kind == "dangling symlink" {
					target = filepath.Join(root, "missing")
				}
				if err := os.Symlink(target, profileDir); err != nil {
					t.Fatal(err)
				}
			default:
				if err := os.Mkdir(profileDir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(profileDir, "auth.json"), []byte("keep"), 0o600); err != nil {
					t.Fatal(err)
				}
				for name, target := range map[string]string{"sessions": sessionsDir, "archived_sessions": archiveDir} {
					if err := os.Symlink(target, filepath.Join(profileDir, name)); err != nil {
						t.Fatal(err)
					}
				}
			}

			err := Remove(tt.profileName)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Remove() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.daemonFails {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 17 {
					t.Errorf("expected daemon exit code 17, got %v", err)
				}
			}
			log, logErr := os.ReadFile(logPath)
			if tt.wantDaemon {
				want := strings.Join([]string{profileDir, "app-server", "daemon", "stop", ""}, "\n")
				if logErr != nil || string(log) != want {
					t.Errorf("daemon invocation = %q, error = %v; want %q", log, logErr, want)
				}
			} else if !errors.Is(logErr, os.ErrNotExist) {
				t.Errorf("expected no daemon invocation, got %q, error = %v", log, logErr)
			}
			if !tt.wantErr || tt.kind == "missing" {
				if _, err := os.Lstat(profileDir); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("expected profile to be absent, got %v", err)
				}
			} else {
				switch tt.kind {
				case "file":
					preserved = append(preserved, profileDir)
				case "symlink", "dangling symlink":
					want := outsideDir
					if tt.kind == "dangling symlink" {
						want = filepath.Join(root, "missing")
					}
					if got, err := os.Readlink(profileDir); err != nil || got != want {
						t.Errorf("profile symlink = %q, error = %v; want %q", got, err, want)
					}
				default:
					preserved = append(preserved, filepath.Join(profileDir, "auth.json"))
					for name, want := range map[string]string{"sessions": sessionsDir, "archived_sessions": archiveDir} {
						if got, err := os.Readlink(filepath.Join(profileDir, name)); err != nil || got != want {
							t.Errorf("%s link = %q, error = %v; want %q", name, got, err, want)
						}
					}
				}
			}
			for _, path := range preserved {
				if got, err := os.ReadFile(path); err != nil || string(got) != "keep" {
					t.Errorf("preserved file %s = %q, error = %v; want keep", path, got, err)
				}
			}
		})
	}
}
