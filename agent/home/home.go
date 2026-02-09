package home

import (
	"github.com/xiehqing/common/pkg/logs"
	"os"
	"path/filepath"
	"strings"
)

var homedir, homedirErr = os.UserHomeDir()

func init() {
	if homedirErr != nil {
		logs.Errorf("failed to get user home directory，%v", homedirErr)
	}
}

// Dir returns the user's home directory.
func Dir() string {
	return homedir
}

// Short returns the short form of the given path.
func Short(p string) string {
	if homedir == "" || !strings.HasPrefix(p, homedir) {
		return p
	}
	return filepath.Join("~", strings.TrimPrefix(p, homedir))
}

// Long returns the long form of the given path.
func Long(p string) string {
	if homedir == "" || !strings.HasPrefix(p, "~") {
		return p
	}
	return strings.Replace(p, "~", homedir, 1)
}
