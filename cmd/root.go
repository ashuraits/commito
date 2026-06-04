package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexshuraits/commito/internal/server"
	"github.com/alexshuraits/commito/web"
	"github.com/spf13/cobra"
)

var (
	port    int
	noOpen  bool
	repoDir string
)

var rootCmd = &cobra.Command{
	Use:   "commito",
	Short: "Git diff viewer in your browser",
	RunE:  run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 7171, "Port to listen on")
	rootCmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't open browser automatically")
	rootCmd.Flags().StringVarP(&repoDir, "dir", "d", "", "Git repo directory (defaults to current dir)")
}

func run(cmd *cobra.Command, args []string) error {
	if repoDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		repoDir = findGitRoot(dir)
	} else {
		abs, err := filepath.Abs(repoDir)
		if err != nil {
			return err
		}
		repoDir = abs
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		ln, err = net.Listen("tcp", ":0")
		if err != nil {
			return err
		}
	}
	actualPort := ln.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	srv := server.New(repoDir, web.FS)
	httpServer := &http.Server{Handler: srv.Handler()}

	fmt.Printf("commito %s → \033]8;;%s\033\\%s\033]8;;\033\\  (repo: %s)\n", Version, url, url, repoDir)
	checkForUpdate()

	if !noOpen {
		go func() {
			time.Sleep(200 * time.Millisecond)
			openBrowser(url)
		}()
	}

	return httpServer.Serve(ln)
}

func findGitRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
