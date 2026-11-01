package codex

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("Codex helper requires a POSIX shell")
	}
	tests := []struct {
		name    string
		args    []string
		exit    int
		missing bool
	}{
		{name: "no arguments"},
		{name: "preserve arguments", args: []string{"exec", "--", "", "two words", "line\nbreak", "'quoted'", "$(echo injected); *"}},
		{name: "nonzero exit", args: []string{"resume", "session-id"}, exit: 17},
		{name: "executable missing", missing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "profile home")
			bin := filepath.Join(root, "bin")
			logPath := filepath.Join(root, "invocation")
			for _, dir := range []string{home, bin} {
				if err := os.Mkdir(dir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if !tt.missing {
				helper := "#!/bin/sh\nprintf '%s\\000' \"$CODEX_HOME\" \"$@\" > \"$CODEX_PROFILES_TEST_LOG\"\nexit \"$CODEX_PROFILES_TEST_EXIT\"\n"
				if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(helper), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("CODEX_HOME", filepath.Join(root, "parent-home"))
			t.Setenv("PATH", bin)
			t.Setenv("CODEX_PROFILES_TEST_LOG", logPath)
			t.Setenv("CODEX_PROFILES_TEST_EXIT", strconv.Itoa(tt.exit))

			err := Run(context.Background(), home, tt.args...)

			switch {
			case tt.missing:
				if !errors.Is(err, exec.ErrNotFound) {
					t.Errorf("expected executable not found, got %v", err)
				}
			case tt.exit != 0:
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != tt.exit {
					t.Errorf("expected exit code %d, got %v", tt.exit, err)
				}
			default:
				if err != nil {
					t.Errorf("Run() error = %v", err)
				}
			}
			log, logErr := os.ReadFile(logPath)
			if tt.missing {
				if !errors.Is(logErr, os.ErrNotExist) {
					t.Errorf("expected no invocation log, got %q, error = %v", log, logErr)
				}
				return
			}
			want := strings.Join(append([]string{home}, tt.args...), "\x00") + "\x00"
			if logErr != nil || string(log) != want {
				t.Errorf("Codex invocation = %q, error = %v; want %q", log, logErr, want)
			}
		})
	}
}
