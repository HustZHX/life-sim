package config

import (
	"os"
	"path/filepath"
)

// ResolveDatabasePath 将相对路径解析为稳定的绝对路径，避免从不同工作目录启动时读写不同库。
func ResolveDatabasePath(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	if env := os.Getenv("LIFE_SIM_DATA_DIR"); env != "" {
		return filepath.Join(env, "lifesim.db")
	}

	// 优先：项目内 backend/data（与 go.mod 位置无关）
	if root := findProjectRoot(); root != "" {
		p := filepath.Join(root, "backend", "data", "lifesim.db")
		if err := os.MkdirAll(filepath.Dir(p), 0755); err == nil {
			return p
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return rel
	}
	return filepath.Join(wd, rel)
}

// ResolveChroniclesDir 返回 data/chronicles/<branch>/ 目录（按当前 git 分支）。
func ResolveChroniclesDir() string {
	branch := CurrentGitBranch()
	root := findProjectRoot()
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	if root == "" {
		root = "."
	}
	p := filepath.Join(root, "data", "chronicles", branch)
	_ = os.MkdirAll(p, 0755)
	return p
}

// ResolveLightNovelsDir 返回 data/light-novels/<branch>/ 目录（按当前 git 分支）。
func ResolveLightNovelsDir() string {
	branch := CurrentGitBranch()
	root := findProjectRoot()
	if root == "" {
		if wd, err := os.Getwd(); err == nil {
			root = wd
		}
	}
	if root == "" {
		root = "."
	}
	p := filepath.Join(root, "data", "light-novels", branch)
	_ = os.MkdirAll(p, 0755)
	return p
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "backend", "go.mod")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "frontend")); err == nil {
			if _, err2 := os.Stat(filepath.Join(dir, "backend")); err2 == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
