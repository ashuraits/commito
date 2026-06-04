package git

import (
	"os/exec"
	"strings"

	"github.com/alexshuraits/commito/internal/models"
)

func GetCommits(repoPath string, limit int) ([]models.Commit, error) {
	args := []string{"-C", repoPath, "log", "--pretty=format:%H\x1f%h\x1f%s\x1f%an\x1f%ar", "-n", "20"}
	if limit > 0 {
		args[len(args)-1] = string(rune('0' + limit/10))
		args = []string{"-C", repoPath, "log", "--pretty=format:%H\x1f%h\x1f%s\x1f%an\x1f%ar", "-n", itoa(limit)}
	}
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return []models.Commit{}, nil // no commits yet
	}

	var commits []models.Commit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\x1f", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, models.Commit{
			Hash:    parts[0],
			Short:   parts[1],
			Message: parts[2],
			Author:  parts[3],
			Date:    parts[4],
		})
	}
	return commits, nil
}

func GetCommitDiff(repoPath, hash string, contextLines int) ([]models.DiffFile, error) {
	args := []string{"-C", repoPath, "show", "--format=", "-U3"}
	if contextLines > 0 {
		args[len(args)-1] = "-U" + itoa(contextLines)
	}
	args = append(args, hash)
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := parseDiff(string(out))
	if files == nil {
		files = []models.DiffFile{}
	}
	return files, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
