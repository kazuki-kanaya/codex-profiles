package profile

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestList(t *testing.T) {
	tests := []struct {
		name        string
		baseMissing bool
		baseFile    bool
		dirs        []string
		files       []string
		symlinks    []string
		want        []string
		wantErr     bool
	}{
		{
			name:        "missing base directory",
			baseMissing: true,
		},
		{
			name: "empty base directory",
		},
		{
			name: "profiles in alphabetical order",
			dirs: []string{"work-2", "sub", "main"},
			want: []string{"main", "sub", "work-2"},
		},
		{
			name:     "exclude invalid names files and symlinks",
			dirs:     []string{"main", "work-2", "Upper", "1main", ".hidden", "with space"},
			files:    []string{"config", "auth.json"},
			symlinks: []string{"linked"},
			want:     []string{"main", "work-2"},
		},
		{
			name:     "base path is a regular file",
			baseFile: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			baseDir := filepath.Join(root, "profiles")
			t.Setenv("CODEX_PROFILES_HOME", baseDir)

			switch {
			case tt.baseMissing:
			case tt.baseFile:
				if err := os.WriteFile(baseDir, []byte("not a directory"), 0o600); err != nil {
					t.Fatal(err)
				}
			default:
				if err := os.Mkdir(baseDir, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range tt.dirs {
				if err := os.Mkdir(filepath.Join(baseDir, name), 0o700); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join(baseDir, name), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range tt.symlinks {
				if err := os.Symlink(root, filepath.Join(baseDir, name)); err != nil {
					t.Fatal(err)
				}
			}

			got, err := List()

			if (err != nil) != tt.wantErr {
				t.Fatalf("List() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("List() = %q, want %q", got, tt.want)
			}
		})
	}
}
