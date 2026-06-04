package git

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/alexshuraits/commito/internal/models"
)

func SearchFiles(repoPath, query string) ([]models.SearchResult, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var results []models.SearchResult
	q := strings.ToLower(query)
	for _, path := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if path == "" {
			continue
		}
		if strings.Contains(strings.ToLower(path), q) {
			results = append(results, models.SearchResult{
				Path: path,
				Type: "file",
			})
		}
		if len(results) >= 50 {
			break
		}
	}
	return results, nil
}

func SearchContent(repoPath, query string) ([]models.SearchResult, error) {
	cmd := exec.Command("git", "-C", repoPath, "grep", "-n", "-i", "--", query)
	out, _ := cmd.Output() // grep returns exit 1 when no matches, ignore error

	var results []models.SearchResult
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}
		lineNo, _ := strconv.Atoi(parts[1])
		results = append(results, models.SearchResult{
			Path:    parts[0],
			Line:    lineNo,
			Content: parts[2],
			Type:    "content",
		})
		if len(results) >= 50 {
			break
		}
	}
	return results, nil
}
