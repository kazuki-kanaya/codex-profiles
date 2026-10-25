package profile

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "single letter", input: "a"},
		{name: "lowercase", input: "main"},
		{name: "digits", input: "work2"},
		{name: "hyphen", input: "work-main"},
		{name: "underscore", input: "work_main"},
		{name: "empty", input: "", wantErr: true},
		{name: "uppercase", input: "Main", wantErr: true},
		{name: "starts with digit", input: "1main", wantErr: true},
		{name: "starts with hyphen", input: "-main", wantErr: true},
		{name: "starts with underscore", input: "_main", wantErr: true},
		{name: "space", input: "my account", wantErr: true},
		{name: "slash", input: "work/main", wantErr: true},
		{name: "path traversal", input: "../main", wantErr: true},
		{name: "non-ASCII", input: "仕事用", wantErr: true},
		{name: "trailing newline", input: "main\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ValidateName(%q) error = %v, wantErr = %v",
					tt.input,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
