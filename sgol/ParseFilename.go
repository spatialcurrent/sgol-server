package sgol

import (
	"path/filepath"
	"strings"
)

func ParseFilename(filename string, lower bool) (string, string) {
	if strings.Contains(filename, ".") {
		ext := filepath.Ext(filename)
		if len(ext) > 0 {
			ext = ext[1:]
		}
		if lower {
			return strings.TrimSuffix(filename, "."+ext), strings.ToLower(ext)
		} else {
			return strings.TrimSuffix(filename, "."+ext), ext
		}
	} else {
		return filename, ""
	}
}
