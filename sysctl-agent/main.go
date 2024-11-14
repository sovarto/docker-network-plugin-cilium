package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/cilium/cilium/pkg/datapath/linux/sysctl"
	"github.com/cilium/cilium/pkg/datapath/tables"
	"github.com/spf13/afero"
)

// SysctlHandler handles sysctl requests
type SysctlHandler struct {
	sysctl sysctl.Sysctl
}

// NewSysctlHandler creates a new SysctlHandler
func NewSysctlHandler(sysctl sysctl.Sysctl) *SysctlHandler {
	return &SysctlHandler{
		sysctl: sysctl,
	}
}

// SysctlRequest represents a sysctl request
type SysctlRequest struct {
	Name     []string         `json:"name"`
	Val      interface{}      `json:"val,omitempty"`
	Settings []tables.Sysctl  `json:"settings,omitempty"`
}

// SysctlResponse represents a sysctl response
type SysctlResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Disable handles the Disable sysctl operation
func (h *SysctlHandler) Disable(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err := h.sysctl.Disable(req.Name)
	respond(w, nil, err)
}

// Enable handles the Enable sysctl operation
func (h *SysctlHandler) Enable(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err := h.sysctl.Enable(req.Name)
	respond(w, nil, err)
}

// Write handles the Write sysctl operation
func (h *SysctlHandler) Write(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	val, ok := req.Val.(string)
	if !ok {
		http.Error(w, "Invalid value for 'val', expected string", http.StatusBadRequest)
		return
	}
	err := h.sysctl.Write(req.Name, val)
	respond(w, nil, err)
}

// WriteInt handles the WriteInt sysctl operation
func (h *SysctlHandler) WriteInt(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// JSON numbers are decoded as float64
	valFloat, ok := req.Val.(float64)
	if !ok {
		http.Error(w, "Invalid value for 'val', expected number", http.StatusBadRequest)
		return
	}
	val := int64(valFloat)
	err := h.sysctl.WriteInt(req.Name, val)
	respond(w, nil, err)
}

// ApplySettings handles the ApplySettings sysctl operation
func (h *SysctlHandler) ApplySettings(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	err := h.sysctl.ApplySettings(req.Settings)
	respond(w, nil, err)
}

// Read handles the Read sysctl operation
func (h *SysctlHandler) Read(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	val, err := h.sysctl.Read(req.Name)
	respond(w, val, err)
}

// ReadInt handles the ReadInt sysctl operation
func (h *SysctlHandler) ReadInt(w http.ResponseWriter, r *http.Request) {
	var req SysctlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	val, err := h.sysctl.ReadInt(req.Name)
	respond(w, val, err)
}

// Helper function to respond to the client
func respond(w http.ResponseWriter, result interface{}, err error) {
	resp := SysctlResponse{
		Result: result,
	}
	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	fs := afero.NewOsFs()
	sysctlInterface := sysctl.NewDirectSysctl(fs, "/proc")
	handler := NewSysctlHandler(sysctlInterface)

	// Define the Unix socket path
	socketPath := "/var/run/cilium/sysctl_service.sock"

	// Remove any existing socket
	os.Remove(socketPath)

	// Listen on the Unix socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen on socket: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Set socket permissions to restrict access
	if err := os.Chmod(socketPath, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set socket permissions: %v\n", err)
		os.Exit(1)
	}

	// Handle HTTP routes
	http.HandleFunc("/sysctl/disable", handler.Disable)
	http.HandleFunc("/sysctl/enable", handler.Enable)
	http.HandleFunc("/sysctl/write", handler.Write)
	http.HandleFunc("/sysctl/writeInt", handler.WriteInt)
	http.HandleFunc("/sysctl/applySettings", handler.ApplySettings)
	http.HandleFunc("/sysctl/read", handler.Read)
	http.HandleFunc("/sysctl/readInt", handler.ReadInt)

	fmt.Printf("Sysctl service is listening on Unix socket: %s\n", socketPath)
	if err := http.Serve(listener, nil); err != nil {
		fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		os.Exit(1)
	}
}
