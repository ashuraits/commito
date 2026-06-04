package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alexshuraits/commito/internal/models"
)

func GetStatus(repoPath string) ([]models.StatusFile, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain=v1", "-u")
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

type FilesResponse struct {
	Tracked []string `json:"tracked"`
	Ignored []string `json:"ignored"`
}

func ListAllFiles(repoPath string) (*FilesResponse, error) {
	trackedCmd := exec.Command("git", "-C", repoPath, "ls-files")
	trackedOut, err := trackedCmd.Output()
	if err != nil {
		return nil, err
	}
	tracked := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(trackedOut)), "\n") {
		if line != "" {
			tracked = append(tracked, line)
		}
	}

	ignoredCmd := exec.Command("git", "-C", repoPath, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory")
	ignoredOut, _ := ignoredCmd.Output()
	ignored := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(ignoredOut)), "\n") {
		if line != "" {
			ignored = append(ignored, line)
		}
	}

	return &FilesResponse{Tracked: tracked, Ignored: ignored}, nil
}

type DirEntry struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func ListDir(repoPath, dirPath string) ([]DirEntry, error) {
	entries, err := os.ReadDir(filepath.Join(repoPath, dirPath))
	if err != nil {
		return nil, err
	}
	result := []DirEntry{}
	for _, e := range entries {
		p := dirPath + "/" + e.Name()
		result = append(result, DirEntry{Path: p, IsDir: e.IsDir()})
	}
	return result, nil
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
