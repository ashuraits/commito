package git

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func WriteFile(repoPath, filePath, content string) error {
	full := filepath.Join(repoPath, filePath)
	return os.WriteFile(full, []byte(content), 0644)
}

func DeleteFile(repoPath, filePath string) error {
	full := filepath.Join(repoPath, filePath)
	if !strings.HasPrefix(full, repoPath) {
		return os.ErrInvalid
	}
	return os.RemoveAll(full)
}

func RenameFile(repoPath, from, to string) error {
	src := filepath.Join(repoPath, from)
	dst := filepath.Join(repoPath, to)
	if !strings.HasPrefix(src, repoPath) || !strings.HasPrefix(dst, repoPath) {
		return os.ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

func CopyFile(repoPath, from, to string) error {
	src := filepath.Join(repoPath, from)
	dst := filepath.Join(repoPath, to)
	if !strings.HasPrefix(src, repoPath) || !strings.HasPrefix(dst, repoPath) {
		return os.ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
