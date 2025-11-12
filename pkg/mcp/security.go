package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecurityConfig defines security constraints for MCP server
type SecurityConfig struct {
	// AllowedPaths restricts file operations to specific directories
	AllowedPaths []string
	
	// AllowWrite controls whether write operations are permitted
	AllowWrite bool
	
	// MaxFileSize limits the size of files that can be read (in bytes)
	MaxFileSize int64
	
	// RequireConfirmation requires explicit user approval for destructive operations
	RequireConfirmation bool
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() *SecurityConfig {
	cwd, _ := os.Getwd()
	return &SecurityConfig{
		AllowedPaths:        []string{cwd},
		AllowWrite:          false, // Read-only by default!
		MaxFileSize:         10 * 1024 * 1024, // 10MB max
		RequireConfirmation: true,
	}
}

// ValidatePath checks if a file path is allowed by security policy
func (sc *SecurityConfig) ValidatePath(filePath string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	
	// Clean the path to prevent traversal attacks
	absPath = filepath.Clean(absPath)
	
	// Check for path traversal attempts
	if strings.Contains(absPath, "..") {
		return fmt.Errorf("path traversal detected: %s", filePath)
	}
	
	// Verify path is within allowed directories
	allowed := false
	for _, allowedPath := range sc.AllowedPaths {
		absAllowed, err := filepath.Abs(allowedPath)
		if err != nil {
			continue
		}
		
		// Check if file is within allowed path
		if strings.HasPrefix(absPath, absAllowed) {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return fmt.Errorf("access denied: %s is outside allowed paths", absPath)
	}
	
	return nil
}

// ValidateFileSize checks if a file is within size limits
func (sc *SecurityConfig) ValidateFileSize(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}
	
	if info.Size() > sc.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d)", info.Size(), sc.MaxFileSize)
	}
	
	return nil
}

// CanWrite checks if write operations are permitted
func (sc *SecurityConfig) CanWrite() error {
	if !sc.AllowWrite {
		return fmt.Errorf("write operations are disabled for security")
	}
	return nil
}

// SanitizeError removes sensitive information from error messages
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}
	
	msg := err.Error()
	
	// Remove absolute paths
	if homeDir, err := os.UserHomeDir(); err == nil {
		msg = strings.ReplaceAll(msg, homeDir, "~")
	}
	
	// Remove username
	if user := os.Getenv("USER"); user != "" {
		msg = strings.ReplaceAll(msg, user, "<user>")
	}
	
	return fmt.Errorf("%s", msg)
}

// AuditLog represents a security audit log entry
type AuditLog struct {
	Timestamp string
	Tool      string
	Action    string
	Path      string
	Allowed   bool
	Reason    string
}

// SecurityAuditor logs security-relevant operations
type SecurityAuditor struct {
	logs []AuditLog
}

func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{
		logs: make([]AuditLog, 0),
	}
}

func (sa *SecurityAuditor) Log(tool, action, path string, allowed bool, reason string) {
	// In production, this should write to a proper audit log
	entry := AuditLog{
		Timestamp: fmt.Sprintf("%d", os.Getpid()), // Simplified for example
		Tool:      tool,
		Action:    action,
		Path:      path,
		Allowed:   allowed,
		Reason:    reason,
	}
	sa.logs = append(sa.logs, entry)
}

func (sa *SecurityAuditor) GetLogs() []AuditLog {
	return sa.logs
}
