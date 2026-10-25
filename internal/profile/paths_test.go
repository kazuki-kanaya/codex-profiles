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
