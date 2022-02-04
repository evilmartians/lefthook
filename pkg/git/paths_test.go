package git

import (
	"os/exec"
	"strings"
	"testing"
)

func BenchmarkRootPath(b *testing.B) {
	repo, _ := NewGit2GoRepository()
	for n := 0; n < b.N; n++ {
		_ = repo.RootPath()
	}
}

func BenchmarkGitPath(b *testing.B) {
	repo, _ := NewGit2GoRepository()
	for n := 0; n < b.N; n++ {
		_ = repo.GitPath()
	}
}

func BenchmarkHooksPath(b *testing.B) {
	repo, _ := NewGit2GoRepository()
	for n := 0; n < b.N; n++ {
		_, _ = repo.HooksPath()
	}
}

func BenchmarkRootPathOld(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		outputBytes, _ := cmd.CombinedOutput()
		_ = strings.TrimSpace(string(outputBytes))
	}
}

func BenchmarkGitPathOld(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("git", "rev-parse", "--git-dir") // that may be relative
		outputBytes, _ := cmd.CombinedOutput()
		_ = strings.TrimSpace(string(outputBytes))
	}
}

func BenchmarkHooksPathOld(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("git", "rev-parse", "--git-path", "hooks")
		_, _ = cmd.CombinedOutput()
	}
}
