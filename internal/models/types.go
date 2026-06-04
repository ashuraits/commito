package models

type DiffFile struct {
	Path     string `json:"path"`
	OldPath  string `json:"oldPath,omitempty"`
	Status   string `json:"status"` // modified, added, deleted, renamed, untracked
	Staged   bool   `json:"staged"`
	Hunks    []Hunk `json:"hunks"`
	AddCount int    `json:"addCount"`
	DelCount int    `json:"delCount"`
}

type Hunk struct {
	Header   string `json:"header"`
	OldStart int    `json:"oldStart"`
	OldCount int    `json:"oldCount"`
	NewStart int    `json:"newStart"`
	NewCount int    `json:"newCount"`
	Lines    []Line `json:"lines"`
}

type Line struct {
	Type    string `json:"type"` // add, del, context
	Content string `json:"content"`
	OldNo   int    `json:"oldNo,omitempty"`
	NewNo   int    `json:"newNo,omitempty"`
}

type StatusFile struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Staged bool   `json:"staged"`
}

type SearchResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line,omitempty"`
	Content string `json:"content,omitempty"`
	Type    string `json:"type"` // file, content
}
