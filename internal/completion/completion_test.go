package completion

import (
	"bytes"
	"strings"
	"testing"
)

func Test_Generate(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		expectedSub string
	}{
		{
			name:        "bash completion",
			args:        []string{"bash"},
			wantErr:     false,
			expectedSub: "bash completion for tskr",
		},
		{
			name:        "zsh completion",
			args:        []string{"zsh"},
			wantErr:     false,
			expectedSub: "#compdef tskr",
		},
		{
			name:        "fish completion",
			args:        []string{"fish"},
			wantErr:     false,
			expectedSub: "# fish completion for tskr",
		},
		{
			name:    "unsupported shell",
			args:    []string{"powershell"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := Generate(tt.args, &out)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !strings.Contains(out.String(), tt.expectedSub) {
				t.Errorf("stdout = %q, expected substring %q", out.String(), tt.expectedSub)
			}
		})
	}
}
