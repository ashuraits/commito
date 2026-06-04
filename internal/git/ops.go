package git

import "os/exec"

func StageFile(repoPath, path string) error {
	return exec.Command("git", "-C", repoPath, "add", "--", path).Run()
}

func UnstageFile(repoPath, path string) error {
	return exec.Command("git", "-C", repoPath, "restore", "--staged", "--", path).Run()
}

func RevertFile(repoPath, path string) error {
	return exec.Command("git", "-C", repoPath, "restore", "--", path).Run()
}

func StageAll(repoPath string) error {
	return exec.Command("git", "-C", repoPath, "add", ".").Run()
}

func UnstageAll(repoPath string) error {
	return exec.Command("git", "-C", repoPath, "restore", "--staged", ".").Run()
}

func RevertAll(repoPath string) error {
	return exec.Command("git", "-C", repoPath, "restore", ".").Run()
}
