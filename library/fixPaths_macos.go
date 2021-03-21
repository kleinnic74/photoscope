package library

import (
	"os"
	"strings"
)

func isPathConversionNeeded(path string) bool {
	return strings.Contains(path, "\\")
}

func convertPath(path string) string {
	return strings.ReplaceAll(path, "\\", string(os.PathSeparator))
}
