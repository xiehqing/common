package fsext

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookup(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	targetFile := filepath.Join(tempDir, "target.txt")
	err := os.WriteFile(targetFile, []byte("test"), 0o644)
	if err != nil {
		t.Fatalf("failed to create target file: %v", err)
	}
	found, err := Lookup(tempDir, "target.txt")
	if err != nil {
		t.Fatalf("failed to lookup target file: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected 1 file, got %d", len(found))
	}
	t.Logf("found: %s", found)
	foundPath, foundA := LookupClosest(tempDir, "target.txt")
	t.Logf("foundPath: %s", foundPath)
	t.Logf("foundA: %v", foundA)

}
