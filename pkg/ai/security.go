package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecurityConfig holds security settings
type SecurityConfig struct {
	AllowedPaths   []string // Directories that can be accessed
	MaxFileSize    int64    // Max file size to read (bytes)
	DisableWrite   bool     // Disable write operations
	WorkingDirOnly bool     // Only allow access within working directory
}

// DefaultSecurityConfig returns secure defaults
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedPaths:   []string{},
		MaxFileSize:    10 * 1024 * 1024, // 10MB
		DisableWrite:   false,
		WorkingDirOnly: true,
	}
}

// ValidatePath checks if a file path is safe to access
func (sc *SecurityConfig) ValidatePath(path string) error {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if we should restrict to working directory
	if sc.WorkingDirOnly {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot get working directory: %w", err)
		}

		cwdAbs, err := filepath.Abs(cwd)
		if err != nil {
			return fmt.Errorf("cannot resolve working directory: %w", err)
		}

		if !strings.HasPrefix(absPath, cwdAbs) {
			return fmt.Errorf("access denied: path '%s' is outside working directory", path)
		}
	}

	// Check allowed paths if specified
	if len(sc.AllowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range sc.AllowedPaths {
			allowedAbs, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, allowedAbs) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied: path '%s' not in allowed directories", path)
		}
	}

	// Check if file exists and get size
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check file size
	if info.Size() > sc.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d bytes)", info.Size(), sc.MaxFileSize)
	}

	return nil
}

// ValidateWriteOperation checks if write operations are allowed
func (sc *SecurityConfig) ValidateWriteOperation(operation string) error {
	if sc.DisableWrite {
		return fmt.Errorf("write operation '%s' is disabled by security policy", operation)
	}
	return nil
}
