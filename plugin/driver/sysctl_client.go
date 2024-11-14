package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cilium/cilium/pkg/datapath/tables"
)

// SysctlClient implements the sysctl.Sysctl interface
type SysctlClient struct {
	httpClient *http.Client
}

// NewSysctlClient creates a new SysctlClient
func NewSysctlClient(socketPath string) *SysctlClient {
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
	return &SysctlClient{
		httpClient: client,
	}
}

// SysctlRequest represents a sysctl request
type SysctlRequest struct {
	Name     []string        `json:"name"`
	Val      interface{}     `json:"val,omitempty"`
	Settings []tables.Sysctl `json:"settings,omitempty"`
}

// SysctlResponse represents a sysctl response
type SysctlResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Disable disables the given sysctl parameter
func (s *SysctlClient) Disable(name []string) error {
	req := SysctlRequest{
		Name: name,
	}
	return doPost(s, "/sysctl/disable", req, nil)
}

// Enable enables the given sysctl parameter
func (s *SysctlClient) Enable(name []string) error {
	req := SysctlRequest{
		Name: name,
	}
	return doPost(s, "/sysctl/enable", req, nil)
}

// Write writes the given sysctl parameter with a string value
func (s *SysctlClient) Write(name []string, val string) error {
	req := SysctlRequest{
		Name: name,
		Val:  val,
	}
	return doPost(s, "/sysctl/write", req, nil)
}

// WriteInt writes the given sysctl parameter with an integer value
func (s *SysctlClient) WriteInt(name []string, val int64) error {
	req := SysctlRequest{
		Name: name,
		Val:  val,
	}
	return doPost(s, "/sysctl/writeInt", req, nil)
}

// ApplySettings applies a list of sysctl settings
func (s *SysctlClient) ApplySettings(sysSettings []tables.Sysctl) error {
	req := SysctlRequest{
		Settings: sysSettings,
	}
	return doPost(s, "/sysctl/applySettings", req, nil)
}

// Read reads the given sysctl parameter and returns its string value
func (s *SysctlClient) Read(name []string) (string, error) {
	req := SysctlRequest{
		Name: name,
	}
	var resp SysctlResponse
	err := doPost(s, "/sysctl/read", req, &resp)
	if err != nil {
		return "", err
	}
	val, ok := resp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected response type: %T", resp.Result)
	}
	return val, nil
}

// ReadInt reads the given sysctl parameter and returns its integer value
func (s *SysctlClient) ReadInt(name []string) (int64, error) {
	req := SysctlRequest{
		Name: name,
	}
	var resp SysctlResponse
	err := doPost(s, "/sysctl/readInt", req, &resp)
	if err != nil {
		return 0, err
	}
	// JSON numbers are decoded as float64
	valFloat, ok := resp.Result.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected response type: %T", resp.Result)
	}
	return int64(valFloat), nil
}

func (s *SysctlClient) Healthcheck() error {
	_, err := s.httpClient.Get("http://unix/health")
	return err
}

// Helper method to send POST requests
func doPost(s *SysctlClient, path string, reqBody SysctlRequest, respBody *SysctlResponse) error {
	url := "http://unix" + path

	// Serialize the request body to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to serialize request: %v", err)
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and decode the response
	if respBody != nil {
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return fmt.Errorf("failed to decode response: %v", err)
		}
		// Check for error in response
		if respBody.Error != "" {
			return fmt.Errorf("server error: %s", respBody.Error)
		}
	} else {
		// For methods that don't expect a response body
		if resp.StatusCode != http.StatusOK {
			var errorResp SysctlResponse
			if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
				return fmt.Errorf("failed to decode error response: %v", err)
			}
			return fmt.Errorf("server error: %s", errorResp.Error)
		}
	}

	return nil
}
