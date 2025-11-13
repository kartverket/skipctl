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

	// Check if file/directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	// Check file size (only for files, not directories)
	if !info.IsDir() && info.Size() > sc.MaxFileSize {
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

// ValidatePathForWrite validates both path access and write permissions
func (sc *SecurityConfig) ValidatePathForWrite(path string, operation string) error {
	// First validate path access
	if err := sc.ValidatePath(path); err != nil {
		return err
	}

	// Then validate write operation is allowed
	return sc.ValidateWriteOperation(operation)
}

// IsSafePath performs basic path safety checks
func IsSafePath(path string) bool {
	// Check for suspicious patterns
	suspicious := []string{
		"..",        // Directory traversal
		"~",         // Home directory expansion
		"/etc/",     // System files
		"/var/",     // System files
		"/usr/",     // System files
		"/bin/",     // System binaries
		"/sbin/",    // System binaries
		"/System/",  // macOS system files
		"/Library/", // macOS system libraries
	}

	for _, pattern := range suspicious {
		if strings.Contains(path, pattern) {
			return false
		}
	}

	return true
}

// NewStrictSecurityConfig returns a very restrictive security config
func NewStrictSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedPaths:   []string{},
		MaxFileSize:    1 * 1024 * 1024, // 1MB
		DisableWrite:   true,            // Read-only
		WorkingDirOnly: true,
	}
}

// NewRelaxedSecurityConfig returns a more permissive config for development
func NewRelaxedSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedPaths:   []string{},
		MaxFileSize:    50 * 1024 * 1024, // 50MB
		DisableWrite:   false,
		WorkingDirOnly: false, // Allow access outside working dir
	}
}
