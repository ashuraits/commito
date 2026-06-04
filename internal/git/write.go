package git

import (
	"os"
	"path/filepath"
)

func WriteFile(repoPath, filePath, content string) error {
	full := filepath.Join(repoPath, filePath)
	return os.WriteFile(full, []byte(content), 0644)
}
