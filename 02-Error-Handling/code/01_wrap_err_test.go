package code

import (
|	"fmt"
	"io"
	"os"
	"testing"

	"github.com/pkg/errors"
)

// TestWrapError
func TestWrapError(t *testing.T) {
	_, err := ReadFile("test")
	fmt.Printf("err:%+v", err)
	fmt.Println("err:", err)
}

func ReadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		// Wrapf err： 包含堆栈信息，可以格式化错误内容
		return nil, errors.Wrapf(err, "failed to open %q", path) // %q 单引号围绕的字符字面值，由Go语法安全地转义，这样中文文件名也能正确显示
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		// Wraperr： 包含堆栈信息
		return nil, errors.Wrap(err, "read failed")
	}
	return buf, nil
}
