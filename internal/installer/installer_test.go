package installer

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_FilepathInPATH(t *testing.T) {
	pathEnv := "/usr/bin:/bin:/usr/local/bin"
	if !FilepathInPATH("/usr/bin", pathEnv) {
		t.Errorf("expected /usr/bin to be in PATH")
	}
	if FilepathInPATH("/nonexistent/dir", pathEnv) {
		t.Errorf("did not expect /nonexistent/dir to be in PATH")
	}
}

func Test_Install(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "bin")

	var stdout bytes.Buffer
	err := Install([]string{"--dir", targetDir}, &stdout)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "tskr installed to") {
		t.Errorf("stdout = %q, expected 'tskr installed to'", stdout.String())
	}

	installedBin := filepath.Join(targetDir, "tskr")
	if _, err := os.Stat(installedBin); err != nil {
		t.Errorf("expected installed binary %s to exist", installedBin)
	}
}
