package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestExecuteExitCode(t *testing.T) {
	if os.Getenv("CODEX_PROFILES_TEST_EXECUTE") == "1" {
		rootCmd.SetArgs([]string{os.Getenv("CODEX_PROFILES_TEST_PROFILE")})
		Execute()
		os.Exit(0)
	}
	if runtime.GOOS == "windows" {
		t.Skip("fake codex requires a POSIX shell")
	}
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "main"), 0o700); err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(binDir, "codex"), []byte("#!/bin/sh\nexit \"$CODEX_PROFILES_TEST_EXIT\"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_PROFILES_HOME", root)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CODEX_PROFILES_TEST_EXECUTE", "1")

	for _, test := range []struct {
		name    string
		profile string
		exit    int
		want    int
	}{
		{"success", "main", 0, 0},
		{"codex failure", "main", 37, 37},
		{"launcher failure", "missing", 0, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("CODEX_PROFILES_TEST_PROFILE", test.profile)
			t.Setenv("CODEX_PROFILES_TEST_EXIT", strconv.Itoa(test.exit))
			cmd := exec.Command(os.Args[0], "-test.run=^TestExecuteExitCode$")
			err := cmd.Run()
			got := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				got = exitErr.ExitCode()
			}
			if got != test.want {
				t.Fatalf("exit code = %d, want %d", got, test.want)
			}
		})
	}
}
