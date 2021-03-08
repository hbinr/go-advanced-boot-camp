package code

import (
	"io"
	"os"
	"testing"

	"github.com/pkg/errors"
)

// TestWrapError
func TestWrapError(t *testing.T) {
	_, err := ReadFile("test")
	t.Log(err)
}

func ReadFile(path string) ([]byte, error) {
	f, err := os.Open("filePath")
	if err != nil {
		// Wrap err： 包含堆栈信息
		return nil, errors.Wrap(err, "open failed")
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}
