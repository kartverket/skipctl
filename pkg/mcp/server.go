package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/kartverket/skipctl/pkg/logging"
)

type Server struct {
	reader   io.Reader
	writer   io.Writer
	logger   *slog.Logger
	security *SecurityConfig
	auditor  *SecurityAuditor
}

func NewServer(reader io.Reader, writer io.Writer) *Server {
	return &Server{
		reader:   reader,
		writer:   writer,
		logger:   logging.Logger(),
		security: DefaultSecurityConfig(),
		auditor:  NewSecurityAuditor(),
	}
}

// NewServerWithSecurity creates a server with custom security configuration
func NewServerWithSecurity(reader io.Reader, writer io.Writer, security *SecurityConfig) *Server {
	return &Server{
		reader:   reader,
		writer:   writer,
		logger:   logging.Logger(),
		security: security,
		auditor:  NewSecurityAuditor(),
	}
}

// JSON-RPC 2.0 types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *Server) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.reader)
	encoder := json.NewEncoder(s.writer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					return err
				}
				return nil // EOF
			}

			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var req JSONRPCRequest
			if err := json.Unmarshal(line, &req); err != nil {
				s.logger.Error("Failed to parse request", "error", err)
				continue
			}

			resp := s.handleRequest(ctx, &req)
			if err := encoder.Encode(resp); err != nil {
				return err
			}
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, req *JSONRPCRequest) *JSONRPCResponse {
	if req.JSONRPC != "2.0" {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32600,
				Message: "Invalid Request: jsonrpc must be 2.0",
			},
			ID: req.ID,
		}
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
			ID: req.ID,
		}
	}
}

type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Capabilities    Capabilities `json:"capabilities"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Capabilities struct {
	Tools ToolsCapability `json:"tools"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo: ServerInfo{
				Name:    "skipctl",
				Version: "0.1.0",
			},
			Capabilities: Capabilities{
				Tools: ToolsCapability{
					ListChanged: false,
				},
			},
		},
		ID: req.ID,
	}
}
