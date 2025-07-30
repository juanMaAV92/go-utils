package path

import (
	"os"
	"strings"
)

func GetMainPath() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	idx := strings.LastIndex(path, "/tests")
	if idx < 0 {
		return path
	}
	return path[:idx]
}
