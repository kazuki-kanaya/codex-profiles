package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCancelProfileChanges(t *testing.T) {
	for _, command := range []*cobra.Command{removeCmd, renameCmd} {
		for _, answer := range []string{"n\n", "\n", ""} {
			t.Run(command.Name()+"/"+answer, func(t *testing.T) {
				baseDir := t.TempDir()
				oldDir := filepath.Join(baseDir, "old")
				if err := os.Mkdir(oldDir, 0o700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("CODEX_PROFILES_HOME", baseDir)
				t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", filepath.Join(baseDir, "archive"))
				t.Setenv("PATH", t.TempDir())
				command.SetIn(strings.NewReader(answer))
				command.SetOut(io.Discard)
				t.Cleanup(func() {
					command.SetIn(nil)
					command.SetOut(nil)
				})
				args := []string{"old"}
				if command == renameCmd {
					args = append(args, "new")
				}
				if err := command.RunE(command, args); err != nil {
					t.Fatal(err)
				}
				if info, err := os.Stat(oldDir); err != nil || !info.IsDir() {
					t.Fatalf("original profile not preserved: %v", err)
				}
				if _, err := os.Lstat(filepath.Join(baseDir, "new")); !os.IsNotExist(err) {
					t.Fatalf("unexpected destination: %v", err)
				}
			})
		}
	}
}
