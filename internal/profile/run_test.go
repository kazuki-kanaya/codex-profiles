package profile

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("Codex helper requires a POSIX shell")
	}
	tests := []struct {
		name    string
		profile string
		kind    string
		wantErr bool
	}{
		{name: "existing profile", profile: "main", kind: "directory"},
		{name: "different profile", profile: "work-2", kind: "directory"},
		{name: "missing profile", profile: "main", kind: "missing", wantErr: true},
		{name: "regular file", profile: "main", kind: "file", wantErr: true},
		{name: "symlink", profile: "main", kind: "symlink", wantErr: true},
		{name: "dangling symlink", profile: "main", kind: "dangling symlink", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			baseDir := filepath.Join(root, "profiles")
			home := filepath.Join(baseDir, tt.profile)
			bin := filepath.Join(root, "bin")
			logPath := filepath.Join(root, "invocation")
			for _, dir := range []string{baseDir, bin} {
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			switch tt.kind {
			case "directory":
				if err := os.Mkdir(home, 0o700); err != nil {
					t.Fatal(err)
				}
			case "file":
				if err := os.WriteFile(home, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			case "symlink", "dangling symlink":
				target := root
				if tt.kind == "dangling symlink" {
					target = filepath.Join(root, "missing")
				}
				if err := os.Symlink(target, home); err != nil {
					t.Fatal(err)
				}
			}
			helper := "#!/bin/sh\nprintf '%s\\000' \"$CODEX_HOME\" \"$@\" > \"$CODEX_PROFILES_TEST_LOG\"\n"
			if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(helper), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CODEX_PROFILES_HOME", baseDir)
			t.Setenv("CODEX_HOME", filepath.Join(root, "parent-home"))
			t.Setenv("PATH", bin)
			t.Setenv("CODEX_PROFILES_TEST_LOG", logPath)
			args := []string{"resume", "session-id", "--model", "test-model"}

			err := Run(tt.profile, args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr = %v", err, tt.wantErr)
			}
			log, logErr := os.ReadFile(logPath)
			if tt.wantErr {
				if !errors.Is(logErr, os.ErrNotExist) {
					t.Errorf("expected no Codex invocation, got %q, error = %v", log, logErr)
				}
				return
			}
			want := strings.Join(append([]string{home}, args...), "\x00") + "\x00"
			if logErr != nil || string(log) != want {
				t.Errorf("Codex invocation = %q, error = %v; want %q", log, logErr, want)
			}
		})
	}
}
