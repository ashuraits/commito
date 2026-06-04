package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alexshuraits/commito/internal/models"
)

func GetDiff(repoPath, filePath string, staged bool, contextLines int) (*models.DiffFile, error) {
	args := []string{"-C", repoPath, "diff"}
	if staged {
		args = append(args, "--staged")
	}
	if contextLines > 0 {
		args = append(args, fmt.Sprintf("-U%d", contextLines))
	}
	if filePath != "" {
		args = append(args, "--", filePath)
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := parseDiff(string(out))
	if filePath != "" && len(files) > 0 {
		return &files[0], nil
	}
	if filePath != "" {
		return &models.DiffFile{Path: filePath, Staged: staged}, nil
	}
	if len(files) > 0 {
		return &files[0], nil
	}
	return &models.DiffFile{}, nil
}

func GetAllDiffs(repoPath string, staged bool, contextLines int) ([]models.DiffFile, error) {
	args := []string{"-C", repoPath, "diff"}
	if staged {
		args = append(args, "--staged")
	}
	if contextLines > 0 {
		args = append(args, fmt.Sprintf("-U%d", contextLines))
	}

	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := parseDiff(string(out))
	if files == nil {
		files = []models.DiffFile{}
	}
	for i := range files {
		files[i].Staged = staged
	}
	return files, nil
}

func parseDiff(raw string) []models.DiffFile {
	var files []models.DiffFile
	var current *models.DiffFile
	var currentHunk *models.Hunk
	oldLine, newLine := 0, 0

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			if current != nil {
				if currentHunk != nil {
					current.Hunks = append(current.Hunks, *currentHunk)
					currentHunk = nil
				}
				files = append(files, *current)
			}
			parts := strings.Fields(line)
			path := ""
			if len(parts) >= 4 {
				path = strings.TrimPrefix(parts[3], "b/")
			}
			current = &models.DiffFile{Path: path, Status: "modified"}

		case strings.HasPrefix(line, "new file"):
			if current != nil {
				current.Status = "added"
			}
		case strings.HasPrefix(line, "deleted file"):
			if current != nil {
				current.Status = "deleted"
			}
		case strings.HasPrefix(line, "rename to "):
			if current != nil {
				current.Status = "renamed"
				current.Path = strings.TrimPrefix(line, "rename to ")
			}
		case strings.HasPrefix(line, "rename from "):
			if current != nil {
				current.OldPath = strings.TrimPrefix(line, "rename from ")
			}

		case strings.HasPrefix(line, "@@ "):
			if current == nil {
				continue
			}
			if currentHunk != nil {
				current.Hunks = append(current.Hunks, *currentHunk)
			}
			hunk := parseHunkHeader(line)
			currentHunk = &hunk
			oldLine = hunk.OldStart
			newLine = hunk.NewStart

		case current != nil && currentHunk != nil:
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentHunk.Lines = append(currentHunk.Lines, models.Line{
					Type:    "add",
					Content: line[1:],
					NewNo:   newLine,
				})
				current.AddCount++
				newLine++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentHunk.Lines = append(currentHunk.Lines, models.Line{
					Type:    "del",
					Content: line[1:],
					OldNo:   oldLine,
				})
				current.DelCount++
				oldLine++
			} else if !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "+++") && !strings.HasPrefix(line, "\\") {
				content := line
				if len(line) > 0 {
					content = line[1:]
				}
				currentHunk.Lines = append(currentHunk.Lines, models.Line{
					Type:    "context",
					Content: content,
					OldNo:   oldLine,
					NewNo:   newLine,
				})
				oldLine++
				newLine++
			}
		}
	}

	if current != nil {
		if currentHunk != nil {
			current.Hunks = append(current.Hunks, *currentHunk)
		}
		files = append(files, *current)
	}

	return files
}

func parseHunkHeader(line string) models.Hunk {
	// @@ -old_start,old_count +new_start,new_count @@ ...
	hunk := models.Hunk{Header: line}
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return hunk
	}
	hunk.OldStart, hunk.OldCount = parseRange(parts[1])
	hunk.NewStart, hunk.NewCount = parseRange(parts[2])
	return hunk
}

func parseRange(s string) (start, count int) {
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "+")
	parts := strings.SplitN(s, ",", 2)
	start, _ = strconv.Atoi(parts[0])
	if len(parts) == 2 {
		count, _ = strconv.Atoi(parts[1])
	} else {
		count = 1
	}
	return
}
