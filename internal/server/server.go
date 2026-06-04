package server

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/alexshuraits/commito/internal/git"
)

type Server struct {
	repoPath string
	webFS    fs.FS
}

func New(repoPath string, webFS fs.FS) *Server {
	return &Server{repoPath: repoPath, webFS: webFS}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/stage", s.handleStage)
	mux.HandleFunc("/api/unstage", s.handleUnstage)
	mux.HandleFunc("/api/revert", s.handleRevert)
	mux.HandleFunc("/api/stage-all", s.handleStageAll)
	mux.HandleFunc("/api/unstage-all", s.handleUnstageAll)
	mux.HandleFunc("/api/revert-all", s.handleRevertAll)
	mux.HandleFunc("/api/diff", s.handleDiff)
	mux.HandleFunc("/api/diffs", s.handleAllDiffs)
	mux.HandleFunc("/api/file", s.handleFile)
	mux.HandleFunc("/api/files", s.handleAllFiles)
	mux.HandleFunc("/api/dir", s.handleDir)
	mux.HandleFunc("/api/search/files", s.handleSearchFiles)
	mux.HandleFunc("/api/search/content", s.handleSearchContent)
	mux.HandleFunc("/api/commits", s.handleCommits)
	mux.HandleFunc("/api/commit-diff", s.handleCommitDiff)

	// serve static frontend
	static, _ := fs.Sub(s.webFS, "dist")
	mux.Handle("/", http.FileServer(http.FS(static)))

	return cors(mux)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	branch, _ := git.CurrentBranch(s.repoPath)
	writeJSON(w, map[string]string{"repoPath": s.repoPath, "branch": branch})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	files, err := git.GetStatus(s.repoPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, files)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	staged := r.URL.Query().Get("staged") == "true"
	untracked := r.URL.Query().Get("untracked") == "true"
	ctx, _ := strconv.Atoi(r.URL.Query().Get("context"))
	if ctx == 0 {
		ctx = 3
	}

	diff, err := git.GetDiff(s.repoPath, file, staged, untracked, ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, diff)
}

func (s *Server) handleAllDiffs(w http.ResponseWriter, r *http.Request) {
	staged := r.URL.Query().Get("staged") == "true"
	ctx, _ := strconv.Atoi(r.URL.Query().Get("context"))
	if ctx == 0 {
		ctx = 3
	}

	diffs, err := git.GetAllDiffs(s.repoPath, staged, ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, diffs)
}

func (s *Server) handleFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path required", 400)
		return
	}

	if r.Method == http.MethodPut {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if err := git.WriteFile(s.repoPath, path, string(body)); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(204)
		return
	}

	http.ServeFile(w, r, filepath.Join(s.repoPath, path))
}

func (s *Server) handleAllFiles(w http.ResponseWriter, r *http.Request) {
	files, err := git.ListAllFiles(s.repoPath)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, files)
}

func (s *Server) handleDir(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	entries, err := git.ListDir(s.repoPath, path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, entries)
}

func (s *Server) handleSearchFiles(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results, err := git.SearchFiles(s.repoPath, q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, results)
}

func (s *Server) handleSearchContent(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	results, err := git.SearchContent(s.repoPath, q)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, results)
}

func (s *Server) handleStage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("file")
	if err := git.StageFile(s.repoPath, path); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleUnstage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("file")
	if err := git.UnstageFile(s.repoPath, path); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleRevert(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("file")
	if err := git.RevertFile(s.repoPath, path); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleStageAll(w http.ResponseWriter, r *http.Request) {
	if err := git.StageAll(s.repoPath); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleUnstageAll(w http.ResponseWriter, r *http.Request) {
	if err := git.UnstageAll(s.repoPath); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleRevertAll(w http.ResponseWriter, r *http.Request) {
	if err := git.RevertAll(s.repoPath); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (s *Server) handleCommits(w http.ResponseWriter, r *http.Request) {
	commits, err := git.GetCommits(s.repoPath, 20)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, commits)
}

func (s *Server) handleCommitDiff(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "hash required", 400)
		return
	}
	ctx, _ := strconv.Atoi(r.URL.Query().Get("context"))
	if ctx == 0 {
		ctx = 3
	}
	files, err := git.GetCommitDiff(s.repoPath, hash, ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, files)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
