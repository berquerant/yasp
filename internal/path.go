package internal

import (
	"io/fs"
	"path/filepath"
)

// Find files by filename under the root.
func FindFiles(filename, root string) ([]string, error) {
	var result []string

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == filename {
			result = append(result, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}
