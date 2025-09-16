package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/kartverket/skipctl/pkg/logging"
	"github.com/kartverket/skipctl/pkg/utils"
)

type Document struct {
	Name        string
	Extension   string
	Permissions os.FileMode
	Content     string
	FromStdin   bool
}

func (d *Document) FromPrevHashWithGoGit(hash string) (*Document, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("open git repo: %w", err)
	}

	h, err := repo.ResolveRevision(plumbing.Revision(hash))
	if err != nil {
		return nil, fmt.Errorf("resolve rev %q: %w", hash, err)
	}

	commit, err := repo.CommitObject(*h)
	if err != nil {
		return nil, fmt.Errorf("read commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("read tree: %w", err)
	}

	rel, err := repoRelativePath(d.Name)
	if err != nil {
		return nil, err
	}

	f, err := tree.File(rel)
	if err != nil {
		return nil, fmt.Errorf("file %q not found in commit %s: %w", rel, hash, err)
	}

	content, err := f.Contents()
	if err != nil {
		return nil, fmt.Errorf("read blob: %w", err)
	}

	return &Document{
		Name:        d.Name,
		Extension:   d.Extension,
		Permissions: d.Permissions,
		Content:     content,
		FromStdin:   false,
	}, nil
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

// func (d *Document) FromPrevHash(hash string) (*Document, error) {
// 	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", hash, d.Name))
// 	var out bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd.Stdout = &out
// 	cmd.Stderr = &stderr
// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("unable to read file from hash: file=%s commit_hash=%s", d.Name, hash)
// 	}
// 	return &Document{
// 		Name:        d.Name,
// 		Extension:   d.Extension,
// 		Permissions: d.Permissions,
// 		Content:     out.String(),
// 		FromStdin:   false,
// 	}, nil
// }

func (d *Document) Write(content string) error {
	d.Content = content
	if d.FromStdin {
		if _, err := io.WriteString(os.Stdout, d.Content); err != nil {
			return fmt.Errorf("error writing document to stdout: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(d.Name, []byte(d.Content), d.Permissions); err != nil {
		return fmt.Errorf("error writing document to file: %w", err)
	}
	return nil
}

func FromFiles(filenames []string) ([]*Document, error) {
	var files = []*Document{}

	for _, filename := range filenames {
		fileContent, err := os.ReadFile(filename)

		if err != nil {
			logging.Logger().Error("unable to read file", "filename", filename)
			continue
		}
		finfo, ferr := os.Stat(filename)
		if ferr != nil {
			return nil, fmt.Errorf("unable to get file info %s %w", filename, ferr)
		}
		files = append(files, &Document{
			Name:        filename,
			Extension:   strings.ToLower(filepath.Ext(filename)),
			Permissions: finfo.Mode().Perm(),
			Content:     string(fileContent),
			FromStdin:   false,
		})
	}
	return files, nil
}
func FromStdin() ([]*Document, error) {
	log := logging.Logger()

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	files := bytes.Split(content, []byte("\n---\n"))

	var manifestFiles []*Document

	for i, fileContent := range files {
		ext, detectFiletypeErr := utils.DetectFiletype(fileContent)
		if detectFiletypeErr != nil {
			log.Error(detectFiletypeErr.Error(), "reason", "skipping")
			continue
		}
		manifestFiles = append(manifestFiles, &Document{
			Name:      fmt.Sprintf("stdin_%d", i),
			Extension: ext,
			Content:   string(fileContent),
			FromStdin: true,
		})
	}
	return manifestFiles, nil
}
