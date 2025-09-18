//nolint:revive // the package name is intentional
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GetFileContentFromRef(filename string, ref string) (*string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}

	h, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return nil, fmt.Errorf("resolve ref %q: %w", ref, err)
	}

	commit, err := repo.CommitObject(*h)
	if err != nil {
		return nil, fmt.Errorf("read commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("read tree: %w", err)
	}

	rel, err := repoRelativePath(filename)
	if err != nil {
		return nil, err
	}

	f, err := tree.File(rel)
	if err != nil {
		return nil, fmt.Errorf("file %q not found in commit %s: %w", rel, ref, err)
	}

	content, err := f.Contents()
	if err != nil {
		return nil, fmt.Errorf("read blob: %w", err)
	}

	return &content, nil
}

// repoRelativePath converts an OS path to a repo-relative, slash-separated path.
func repoRelativePath(p string) (string, error) {
	// Make the target absolute so Rel() can compute correctly.
	absP, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}

	root, err := findGitRoot()
	if err != nil {
		return "", fmt.Errorf("find git root: %w", err)
	}
	rel, err := filepath.Rel(root, absP)
	if err != nil {
		return "", fmt.Errorf("rel path: %w", err)
	}

	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("path %q is outside repo root %q", p, root)
	}
	return rel, nil
}

// findGitRoot walks up from CWD to locate the .git directory.
func findGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if abs, fpErr := filepath.Abs(dir); fpErr == nil {
		dir = abs
	}
	for {
		// Accept .git as a dir or a file (worktrees).
		if _, statErr := os.Stat(filepath.Join(dir, ".git")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .git found from %q", dir)
		}
		dir = parent
	}
}
