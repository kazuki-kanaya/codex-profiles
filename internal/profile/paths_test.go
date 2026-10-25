package profile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBaseDir(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("home directory tests cover macOS and Linux")
	}
	home := t.TempDir()
	empty := ""
	absolute := filepath.Join(t.TempDir(), "profiles")
	relative := filepath.Join("custom", "profiles")
	expectedAbsolute, err := filepath.Abs(relative)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		home     string
		override *string
		want     string
		wantErr  bool
	}{
		{
			name: "unset override",
			home: home,
			want: filepath.Join(home, ".codex-profiles"),
		},
		{
			name:     "empty override",
			home:     home,
			override: &empty,
			want:     filepath.Join(home, ".codex-profiles"),
		},
		{
			name:     "absolute override",
			home:     home,
			override: &absolute,
			want:     absolute,
		},
		{
			name:     "relative override",
			home:     home,
			override: &relative,
			want:     expectedAbsolute,
		},
		{
			name:    "missing home",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)
			t.Setenv("CODEX_PROFILES_HOME", "")
			if tt.override == nil {
				if err := os.Unsetenv("CODEX_PROFILES_HOME"); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv("CODEX_PROFILES_HOME", *tt.override)
			}

			got, err := BaseDir()

			if (err != nil) != tt.wantErr {
				t.Fatalf("BaseDir() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("BaseDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSessionsDir(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("home directory tests cover macOS and Linux")
	}

	home := t.TempDir()
	empty := ""
	absolute := filepath.Join(t.TempDir(), "sessions")
	relative := filepath.Join("custom", "sessions")
	expectedAbsolute, err := filepath.Abs(relative)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		home     string
		override *string
		want     string
		wantErr  bool
	}{
		{
			name: "unset override",
			home: home,
			want: filepath.Join(home, ".codex", "sessions"),
		},
		{
			name:     "empty override",
			home:     home,
			override: &empty,
			want:     filepath.Join(home, ".codex", "sessions"),
		},
		{
			name:     "absolute override",
			home:     home,
			override: &absolute,
			want:     absolute,
		},
		{
			name:     "relative override",
			home:     home,
			override: &relative,
			want:     expectedAbsolute,
		},
		{
			name:    "missing home",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)
			t.Setenv("CODEX_PROFILES_HOME", filepath.Join(home, "other-profiles"))
			t.Setenv("CODEX_PROFILES_SESSIONS_DIR", "")
			if tt.override == nil {
				if err := os.Unsetenv("CODEX_PROFILES_SESSIONS_DIR"); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv("CODEX_PROFILES_SESSIONS_DIR", *tt.override)
			}

			got, err := SessionsDir()

			if (err != nil) != tt.wantErr {
				t.Fatalf("SessionsDir() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("SessionsDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArchivedSessionsDir(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("home directory tests cover macOS and Linux")
	}

	home := t.TempDir()
	empty := ""
	absolute := filepath.Join(t.TempDir(), "archived_sessions")
	relative := filepath.Join("custom", "archived_sessions")
	expectedAbsolute, err := filepath.Abs(relative)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		home     string
		override *string
		want     string
		wantErr  bool
	}{
		{
			name: "unset override",
			home: home,
			want: filepath.Join(home, ".codex", "archived_sessions"),
		},
		{
			name:     "empty override",
			home:     home,
			override: &empty,
			want:     filepath.Join(home, ".codex", "archived_sessions"),
		},
		{
			name:     "absolute override",
			home:     home,
			override: &absolute,
			want:     absolute,
		},
		{
			name:     "relative override",
			home:     home,
			override: &relative,
			want:     expectedAbsolute,
		},
		{
			name:    "missing home",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)
			t.Setenv("CODEX_PROFILES_HOME", filepath.Join(home, "other-profiles"))
			t.Setenv("CODEX_PROFILES_SESSIONS_DIR", filepath.Join(home, "other-sessions"))
			t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", "")
			if tt.override == nil {
				if err := os.Unsetenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR"); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv("CODEX_PROFILES_ARCHIVED_SESSIONS_DIR", *tt.override)
			}

			got, err := ArchivedSessionsDir()

			if (err != nil) != tt.wantErr {
				t.Fatalf("ArchivedSessionsDir() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ArchivedSessionsDir() = %q, want %q", got, tt.want)
			}
		})
	}
}
