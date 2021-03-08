package code

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
)

// TestWrapError
func TestErrWithmessage(t *testing.T) {
	_, err := ReadConfig("test")
	t.Log(err)
}

var filePath = "test_file_path"

func ReadConfig(path string) ([]byte, error) {
	home := os.Getenv("HOME")
	config, err := ReadFile(filepath.Join(home, ".yaml"))
	return config, errors.WithMessage(err, "could mot read config")
}
