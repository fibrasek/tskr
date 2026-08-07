package cli

import (
	"bytes"
	"strings"
	"testing"
)

func Test_PrintUsage(t *testing.T) {
	var out bytes.Buffer
	PrintUsage(&out)
	if !strings.Contains(out.String(), "tskr - task reference context manager") {
		t.Errorf("PrintUsage output missing header: %q", out.String())
	}
}

func Test_Run_Help(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	err := Run([]string{"tskr", "--help"}, &out, &errOut)
	if err != nil {
		t.Fatalf("Run(--help) error = %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("Run(--help) missing Usage text: %q", out.String())
	}
}
