package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexshuraits/commito/internal/models"
)

func GetStatus(repoPath string) ([]models.StatusFile, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain=v1")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := []models.StatusFile{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r") // strip \r on Windows
		if len(line) < 4 {
			continue
		}
		x := string(line[0]) // staged status
		y := string(line[1]) // unstaged status
		path := strings.TrimSpace(line[3:])

		// handle renames: "old -> new"
		if strings.Contains(path, " -> ") {
			parts := strings.SplitN(path, " -> ", 2)
			path = parts[1]
		}

		if x != " " && x != "?" {
			files = append(files, models.StatusFile{
				Path:   path,
				Status: statusLabel(x),
				Staged: true,
			})
		}
		if y != " " && y != "?" {
			files = append(files, models.StatusFile{
				Path:   path,
				Status: statusLabel(y),
				Staged: false,
			})
		}
		if x == "?" && y == "?" {
			files = append(files, models.StatusFile{
				Path:   path,
				Status: "untracked",
				Staged: false,
			})
		}
	}
	return files, nil
}

func ListAllFiles(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func ReadFile(repoPath, filePath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoPath, filePath))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func statusLabel(code string) string {
	switch code {
	case "M":
		return "modified"
	case "A":
		return "added"
	case "D":
		return "deleted"
	case "R":
		return "renamed"
	case "C":
		return "copied"
	case "U":
		return "unmerged"
	default:
		return "modified"
	}
}
