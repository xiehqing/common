package home

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	t.Log(filepath.Join(Dir(), ".config", "crush", fmt.Sprintf("%s.json", "crush")))
	t.Log(Dir())
}

func TestShort(t *testing.T) {
	d := filepath.Join(Dir(), "documents", "file.txt")
	t.Log(Short(d))
	ad := filepath.FromSlash("/absolute/path/file.txt")
	t.Log(Short(ad))
}

func TestLong(t *testing.T) {
	d := filepath.Join(Dir(), "documents", "file.txt")
	t.Log(Long(d))
	ad := filepath.FromSlash("/absolute/path/file.txt")
	t.Log(Long(ad))
}
