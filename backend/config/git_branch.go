package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var branchSafeRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// CurrentGitBranch 返回项目根目录下的当前 git 分支名；失败时为 "unknown"。
func CurrentGitBranch() string {
	root := findProjectRoot()
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	if root == "" {
		return "unknown"
	}
	out, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	name := strings.TrimSpace(string(out))
	if name == "" || name == "HEAD" {
		return "unknown"
	}
	return SanitizeBranchDir(name)
}

// SanitizeBranchDir 将分支名转为可作目录名的安全片段。
func SanitizeBranchDir(branch string) string {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "unknown"
	}
	branch = strings.ReplaceAll(branch, string(filepath.Separator), "_")
	branch = branchSafeRe.ReplaceAllString(branch, "_")
	if branch == "" {
		return "unknown"
	}
	return branch
}
