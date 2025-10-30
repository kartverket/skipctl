package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const (
	// ioReadAll buffer sizing and safety cap.
	readBufferGrowSize = 64 << 10 // 65 536 bytes (64 KiB)
	readMaxBytes       = 32 << 20 // 33 554 432 bytes (32 MiB)
)

func GetFileContentAtRef(filename string, ref string) (*string, error) {
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

	content, err := readPathFromTree(repo, tree, rel)
	if err != nil {
		return nil, fmt.Errorf("file %q not found in commit %s: %w", rel, ref, err)
	}
	return &content, nil
}

// readPathFromTree walks the given path starting at "tree" in "repo".
// It detects submodule (gitlink) entries and continues resolution inside the
// submodule repository at the recorded commit. Supports nested submodules.
func readPathFromTree(repo *git.Repository, tree *object.Tree, relPath string) (string, error) {
	parts := splitPath(relPath)
	curRepo := repo
	curTree := tree

	for i, name := range parts {
		entry, err := findEntry(curTree, name)
		if err != nil {
			return "", fmt.Errorf("path %q: %w", join(parts[:i+1]), err)
		}

		switch entry.Mode {
		case filemode.Submodule:
			// Switch into submodule repo at the recorded commit.
			subRepo, subTree, subModErr := openSubmoduleAt(curRepo, name, entry.Hash)
			if subModErr != nil {
				return "", fmt.Errorf("enter submodule %q: %w", join(parts[:i+1]), subModErr)
			}
			curRepo = subRepo
			curTree = subTree

		case filemode.Dir:
			nextTree, treeObjErr := curRepo.TreeObject(entry.Hash)
			if treeObjErr != nil {
				return "", fmt.Errorf("open dir %q: %w", join(parts[:i+1]), treeObjErr)
			}
			curTree = nextTree

		case filemode.Regular, filemode.Executable, filemode.Symlink, filemode.Deprecated:
			// Treat these as file-like entries; only valid if last segment.
			if i == len(parts)-1 {
				return readBlob(curRepo, entry.Hash)
			}
			return "", fmt.Errorf("path %q is not a directory", join(parts[:i+1]))

		case filemode.Empty:
			// Not a valid traversable entry; report clearly.
			return "", fmt.Errorf("path %q refers to an empty entry", join(parts[:i+1]))

		default:
			return "", fmt.Errorf("path %q has unsupported file mode %v", join(parts[:i+1]), entry.Mode)
		}
	}

	return "", fmt.Errorf("path %q resolved to a directory", relPath)
}

// openSubmoduleAt opens the submodule repository located under the current repo’s
// .git/modules/<submoduleName> and returns its tree at the given commit hash.
func openSubmoduleAt(parentRepo *git.Repository, submoduleName string, subCommitHash plumbing.Hash) (*git.Repository, *object.Tree, error) {
	gitDir, err := repoGitDir(parentRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("locate git dir: %w", err)
	}
	// Standard on-disk layout for submodules.
	subGitDir := filepath.Join(gitDir, "modules", submoduleName)
	subRepo, err := git.PlainOpen(subGitDir)
	if err != nil {
		return nil, nil, fmt.Errorf("open submodule repo at %q: %w", subGitDir, err)
	}
	subCommit, err := subRepo.CommitObject(subCommitHash)
	if err != nil {
		return nil, nil, fmt.Errorf("load submodule commit %s: %w", subCommitHash.String(), err)
	}
	subTree, err := subCommit.Tree()
	if err != nil {
		return nil, nil, fmt.Errorf("load submodule tree %s: %w", subCommitHash.String(), err)
	}
	return subRepo, subTree, nil
}

// findEntry returns the tree entry with the given name.
func findEntry(tree *object.Tree, name string) (*object.TreeEntry, error) {
	for i := range tree.Entries {
		if tree.Entries[i].Name == name {
			return &tree.Entries[i], nil
		}
	}
	return nil, fmt.Errorf("entry %q not found", name)
}

func readBlob(repo *git.Repository, h plumbing.Hash) (string, error) {
	blob, err := repo.BlobObject(h)
	if err != nil {
		return "", fmt.Errorf("read blob: %w", err)
	}
	rdr, err := blob.Reader()
	if err != nil {
		return "", fmt.Errorf("blob reader: %w", err)
	}
	defer rdr.Close()
	data, err := ioReadAll(rdr)
	if err != nil {
		return "", fmt.Errorf("read blob data: %w", err)
	}
	return string(data), nil
}

// repoGitDir returns the absolute path to the repository’s git dir,
// resolving the .git file indirection for worktrees if necessary.
func repoGitDir(_ *git.Repository) (string, error) {
	// Resolve from the superproject working directory.
	root, err := findGitRoot()
	if err != nil {
		return "", fmt.Errorf("find git root: %w", err)
	}
	dotGit := filepath.Join(root, ".git")
	fi, err := os.Stat(dotGit)
	if err != nil {
		return "", fmt.Errorf("stat .git: %w", err)
	}
	if fi.IsDir() {
		return dotGit, nil
	}
	// .git is a file; resolve the actual gitdir
	return resolveDotGitFile(dotGit)
}

func resolveDotGitFile(dotGitPath string) (string, error) {
	b, err := os.ReadFile(dotGitPath)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(b))
	const pref = "gitdir: "
	if strings.HasPrefix(strings.ToLower(s), pref) {
		gitdir := strings.TrimSpace(s[len(pref):])
		if !filepath.IsAbs(gitdir) {
			gitdir = filepath.Join(filepath.Dir(dotGitPath), gitdir)
		}
		return gitdir, nil
	}
	return "", fmt.Errorf("%s is not a gitdir file", dotGitPath)
}

func ioReadAll(r io.Reader) ([]byte, error) {
	var b bytes.Buffer
	b.Grow(readBufferGrowSize)
	n, err := b.ReadFrom(io.LimitReader(r, readMaxBytes))
	if err != nil {
		return nil, err
	}
	if n >= readMaxBytes {
		return nil, fmt.Errorf("file exceeds read limit (%d bytes)", readMaxBytes)
	}
	return b.Bytes(), nil
}

// splitPath returns forward-slashed components for a repo-relative path.
func splitPath(p string) []string {
	p = strings.TrimPrefix(filepath.ToSlash(p), "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func join(parts []string) string {
	return strings.Join(parts, "/")
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

// GetFilesystemAt materializes the repository tree at the given git ref to a temporary directory on disk.
// It creates a temporary directory and writes all files from the ref.
// Returns the path to the temporary directory which the caller is responsible for cleaning up.
// Supports submodules using the existing submodule traversal logic.

//nolint:govet // variable shadowing  acceptable
func GetFilesystemAt(ref string) (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return "", fmt.Errorf("open git repo: %w", err)
	}

	h, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return "", fmt.Errorf("resolve ref %q: %w", ref, err)
	}

	commit, err := repo.CommitObject(*h)
	if err != nil {
		return "", fmt.Errorf("read commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("read tree: %w", err)
	}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-ref-*")
	if err != nil {
		return "", fmt.Errorf("create temp directory: %w", err)
	}

	// Write tree to disk
	if err := writeTreeToDisk(repo, tree, tmpDir); err != nil {
		os.RemoveAll(tmpDir) // Clean up on error
		return "", err
	}

	return tmpDir, nil
}

//nolint:govet,exhaustive,mnd, gocognit // variable shadowing, exhaustive switch, magic numbers acceptable for this helper
func writeTreeToDisk(repo *git.Repository, tree *object.Tree, targetDir string) error {
	for i := range tree.Entries {
		entry := &tree.Entries[i]
		targetPath := filepath.Join(targetDir, entry.Name)

		switch entry.Mode {
		case filemode.Dir:
			// Create directory and recurse
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("create directory %q: %w", entry.Name, err)
			}
			subTree, err := repo.TreeObject(entry.Hash)
			if err != nil {
				return fmt.Errorf("read tree for %q: %w", entry.Name, err)
			}
			if err := writeTreeToDisk(repo, subTree, targetPath); err != nil {
				return fmt.Errorf("write tree %q: %w", entry.Name, err)
			}

		case filemode.Submodule:
			// Open submodule and materialize it
			subRepo, subTree, err := openSubmoduleAt(repo, entry.Name, entry.Hash)
			if err != nil {
				return fmt.Errorf("open submodule %q: %w", entry.Name, err)
			}
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("create submodule directory %q: %w", entry.Name, err)
			}
			if err := writeTreeToDisk(subRepo, subTree, targetPath); err != nil {
				return fmt.Errorf("write submodule %q: %w", entry.Name, err)
			}

		case filemode.Regular, filemode.Executable, filemode.Deprecated:
			// Write file content
			content, err := readBlob(repo, entry.Hash)
			if err != nil {
				return fmt.Errorf("read file %q: %w", entry.Name, err)
			}
			mode := os.FileMode(0644)
			if entry.Mode == filemode.Executable {
				mode = 0755
			}
			if err := os.WriteFile(targetPath, []byte(content), mode); err != nil {
				return fmt.Errorf("write file %q: %w", entry.Name, err)
			}

		case filemode.Symlink:
			// Read symlink target and create symlink
			target, err := readBlob(repo, entry.Hash)
			if err != nil {
				return fmt.Errorf("read symlink %q: %w", entry.Name, err)
			}
			if err := os.Symlink(target, targetPath); err != nil {
				return fmt.Errorf("create symlink %q: %w", entry.Name, err)
			}

		default:
			// Skip unsupported file modes
			continue
		}
	}

	return nil
}
