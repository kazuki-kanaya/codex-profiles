package profile

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

type addTestEnv struct {
	home     string
	profile  string
	sessions string
	archive  string
	loginLog string
	ready    string
	sentinel string
}

func setupAddTest(t *testing.T) addTestEnv {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("login helper requires a POSIX shell")
	}
	root := t.TempDir()
	env := addTestEnv{
		home:     filepath.Join(root, "profiles"),
		profile:  filepath.Join(root, "profiles", "main"),
		sessions: filepath.Join(root, "sessions"),
		archive:  filepath.Join(root, "archived_sessions"),
		loginLog: filepath.Join(root, "login.json"),
		ready:    filepath.Join(root, "ready"),
		sentinel: filepath.Join(root, "profiles", "existing", "keep"),
	}
	for _, dir := range []string{env.sessions, env.archive, filepath.Dir(env.sentinel)} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{env.sentinel, filepath.Join(env.sessions, "keep.jsonl"), filepath.Join(env.archive, "keep.jsonl.zst")} {
		if err := os.WriteFile(path, []byte("keep"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	helper := "#!/bin/sh\nCODEX_PROFILES_TEST_MODE=login exec \"$CODEX_PROFILES_TEST_BINARY\" -test.run=^TestAddProcess$ -- \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(helper), 0o700); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", root)
	t.Setenv("CODEX_HOME", filepath.Join(root, "parent-home"))
	t.Setenv("CODEX_PROFILES_HOME", env.home)
	t.Setenv("CODEX_PROFILES_SESSIONS_DIR", env.sessions)
	t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", env.archive)
	t.Setenv("PATH", bin)
	t.Setenv("CODEX_PROFILES_TEST_BINARY", binary)
	t.Setenv("CODEX_PROFILES_TEST_MODE", "")
	t.Setenv("CODEX_PROFILES_TEST_LOGIN", "success")
	t.Setenv("CODEX_PROFILES_TEST_LOG", env.loginLog)
	t.Setenv("CODEX_PROFILES_TEST_READY", env.ready)
	return env
}

func requireAddFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, got, want)
	}
}

func requireAddMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected %s to be absent, got %v", path, err)
	}
}

func requireAddPreserved(t *testing.T, env addTestEnv) {
	t.Helper()
	for _, path := range []string{env.sentinel, filepath.Join(env.sessions, "keep.jsonl"), filepath.Join(env.archive, "keep.jsonl.zst")} {
		requireAddFile(t, path, "keep")
	}
}

func TestAdd(t *testing.T) {
	for _, tt := range []struct {
		name          string
		login         string
		sharedMissing bool
		wantErr       bool
	}{
		{name: "successful login", login: "success"},
		{name: "creates shared directories", login: "success", sharedMissing: true},
		{name: "failed login", login: "failure", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAddTest(t)
			preserved := env
			t.Setenv("CODEX_PROFILES_TEST_LOGIN", tt.login)
			if tt.sharedMissing {
				root := t.TempDir()
				env.sessions = filepath.Join(root, "sessions")
				env.archive = filepath.Join(root, "archived_sessions")
				t.Setenv("CODEX_PROFILES_SESSIONS_DIR", env.sessions)
				t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", env.archive)
			}

			err := Add("main")

			if (err != nil) != tt.wantErr {
				t.Fatalf("Add() error = %v, wantErr = %v", err, tt.wantErr)
			}
			var login struct {
				Home string
				Args []string
			}
			data, readErr := os.ReadFile(env.loginLog)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if err := json.Unmarshal(data, &login); err != nil {
				t.Fatal(err)
			}
			if login.Home != env.profile || len(login.Args) != 1 || login.Args[0] != "login" {
				t.Fatalf("unexpected login invocation: %+v", login)
			}
			if tt.wantErr {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 17 {
					t.Fatalf("expected login exit code 17, got %v", err)
				}
				requireAddMissing(t, env.profile)
			} else {
				requireAddFile(t, filepath.Join(env.profile, "auth.json"), "test credentials")
				requireAddFile(t, filepath.Join(env.profile, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
				for name, target := range map[string]string{"sessions": env.sessions, "archived_sessions": env.archive} {
					got, err := os.Readlink(filepath.Join(env.profile, name))
					if err != nil || got != target {
						t.Fatalf("%s link = %q, error = %v; want %q", name, got, err, target)
					}
					info, err := os.Stat(target)
					if err != nil {
						t.Fatal(err)
					}
					if !info.IsDir() {
						t.Errorf("shared %s is not a directory", name)
					}
				}
				for _, file := range []string{env.profile, filepath.Join(env.profile, "config.toml")} {
					info, err := os.Stat(file)
					if err != nil {
						t.Fatal(err)
					}
					if info.Mode().Perm()&0o077 != 0 {
						t.Errorf("%s is accessible by other users: %o", file, info.Mode().Perm())
					}
				}
			}
			requireAddPreserved(t, preserved)
		})
	}
}

func TestAddRejectsExistingProfile(t *testing.T) {
	for _, kind := range []string{"directory", "file", "symlink", "dangling symlink"} {
		t.Run(kind, func(t *testing.T) {
			env := setupAddTest(t)
			switch kind {
			case "directory":
				if err := os.Mkdir(env.profile, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(env.profile, "auth.json"), []byte("existing credentials"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "file":
				if err := os.WriteFile(env.profile, []byte("existing file"), 0o600); err != nil {
					t.Fatal(err)
				}
			default:
				target := env.sessions
				if kind == "dangling symlink" {
					target = filepath.Join(env.home, "missing")
				}
				if err := os.Symlink(target, env.profile); err != nil {
					t.Fatal(err)
				}
			}

			err := Add("main")

			if err == nil || !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("expected duplicate profile error, got %v", err)
			}
			switch kind {
			case "directory":
				requireAddFile(t, filepath.Join(env.profile, "auth.json"), "existing credentials")
			case "file":
				requireAddFile(t, env.profile, "existing file")
			default:
				want := env.sessions
				if kind == "dangling symlink" {
					want = filepath.Join(env.home, "missing")
				}
				if got, err := os.Readlink(env.profile); err != nil || got != want {
					t.Fatalf("existing link changed: target = %q, error = %v", got, err)
				}
			}
			requireAddMissing(t, env.loginLog)
			requireAddPreserved(t, env)
		})
	}
}

func TestAddPreparationFailure(t *testing.T) {
	for _, target := range []string{"sessions", "archived_sessions", "codex executable"} {
		t.Run(target, func(t *testing.T) {
			env := setupAddTest(t)
			blocked := filepath.Join(t.TempDir(), "blocked")
			if err := os.WriteFile(blocked, []byte("keep"), 0o600); err != nil {
				t.Fatal(err)
			}
			switch target {
			case "sessions":
				t.Setenv("CODEX_PROFILES_SESSIONS_DIR", blocked)
			case "archived_sessions":
				t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", blocked)
			case "codex executable":
				t.Setenv("PATH", t.TempDir())
			}

			if err := Add("main"); err == nil {
				t.Fatal("expected Add to fail")
			}
			requireAddMissing(t, env.profile)
			requireAddMissing(t, env.loginLog)
			requireAddFile(t, blocked, "keep")
			requireAddPreserved(t, env)
		})
	}
}

func TestAddInterrupted(t *testing.T) {
	for _, tt := range []struct {
		name   string
		signal os.Signal
	}{
		{name: "SIGINT", signal: os.Interrupt},
		{name: "SIGTERM", signal: syscall.SIGTERM},
	} {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAddTest(t)
			t.Setenv("CODEX_PROFILES_TEST_LOGIN", "wait")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Getenv("CODEX_PROFILES_TEST_BINARY"), "-test.run=^TestAddProcess$")
			cmd.Env = append(os.Environ(), "CODEX_PROFILES_TEST_MODE=add")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				// Include the login helper if the Add process exits before cleanup.
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				if cmd.ProcessState == nil {
					_ = cmd.Wait()
				}
			}()
			deadline := time.Now().Add(5 * time.Second)
			for {
				if _, err := os.Stat(env.ready); err == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("login helper did not become ready")
				}
				time.Sleep(10 * time.Millisecond)
			}

			if err := cmd.Process.Signal(tt.signal); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Wait(); err != nil {
				t.Fatalf("interrupted Add did not clean up successfully: %v", err)
			}
			requireAddMissing(t, env.profile)
			requireAddPreserved(t, env)
		})
	}
}

func TestAddProcess(t *testing.T) {
	switch os.Getenv("CODEX_PROFILES_TEST_MODE") {
	case "add":
		if err := Add("main"); err == nil {
			t.Fatal("expected interrupted login error")
		}
	case "login":
		index := 0
		for index < len(os.Args) && os.Args[index] != "--" {
			index++
		}
		if index == len(os.Args) {
			t.Fatal("missing helper arguments")
		}
		home := os.Getenv("CODEX_HOME")
		data, err := json.Marshal(struct {
			Home string
			Args []string
		}{home, os.Args[index+1:]})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Getenv("CODEX_PROFILES_TEST_LOG"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		requireAddFile(t, filepath.Join(home, "config.toml"), "cli_auth_credentials_store = \"file\"\n")
		for name, variable := range map[string]string{"sessions": "CODEX_PROFILES_SESSIONS_DIR", "archived_sessions": "CODEX_PROFILES_ARCHIVED_SESSIONS_DIR"} {
			if got, err := os.Readlink(filepath.Join(home, name)); err != nil || got != os.Getenv(variable) {
				t.Fatalf("profile was not ready before login: %s", name)
			}
		}
		if err := os.WriteFile(filepath.Join(home, "auth.json"), []byte("test credentials"), 0o600); err != nil {
			t.Fatal(err)
		}
		switch os.Getenv("CODEX_PROFILES_TEST_LOGIN") {
		case "failure":
			os.Exit(17)
		case "wait":
			if err := os.WriteFile(os.Getenv("CODEX_PROFILES_TEST_READY"), nil, 0o600); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(15 * time.Second)
			for time.Now().Before(deadline) {
				time.Sleep(time.Second)
			}
			t.Fatal("login helper was not stopped")
		}
	default:
		t.Skip("subprocess helper")
	}
}
