package test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Graylog2/go-gelf/gelf"
	"github.com/stretchr/testify/require"
)

// DockerComposeManager manages Docker Compose for integration tests
type DockerComposeManager struct {
	ctx        context.Context
	cancel     context.CancelFunc
	composeCmd *exec.Cmd
}

// NewDockerComposeManager creates a new Docker Compose manager
func NewDockerComposeManager(ctx context.Context) *DockerComposeManager {
	return &DockerComposeManager{
		ctx:    ctx,
		cancel: func() {},
	}
}

// Start starts Graylog via Docker Compose and waits for it to be ready
func (m *DockerComposeManager) Start() error {
	var cancel context.CancelFunc
	m.ctx, cancel = context.WithCancel(context.Background())
	m.cancel = cancel

	// Run docker-compose up
	cmd := exec.CommandContext(m.ctx, "docker-compose", "-f", "docker-compose.test.yml", "up", "-d", "--wait", "--wait-timeout", "60")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start docker compose: %w", err)
	}

	m.composeCmd = cmd

	// Wait for Graylog to be ready
	return m.waitForGraylog()
}

// Stop stops the Docker Compose services
func (m *DockerComposeManager) Stop() error {
	if m.composeCmd != nil {
		m.cancel()
	}
	
	cmd := exec.Command("docker-compose", "-f", "docker-compose.test.yml", "down", "-v", "--remove-orphans")
	return cmd.Run()
}

// waitForGraylog waits until Graylog is accepting connections
func (m *DockerComposeManager) waitForGraylog() error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Graylog")
		case <-ticker.C:
			// Try to connect to GELF UDP port
			conn, err := net.DialTimeout("udp", "localhost:12201", time.Second)
			if err != nil {
				continue
			}
			conn.Close()
			
			// Try HTTP health endpoint
			httpConn, err := net.DialTimeout("tcp", "localhost:9000", time.Second)
			if err != nil {
				continue
			}
			httpConn.Close()
			
			return nil
		}
	}
}

// GELFWriter creates a GELF writer connection to Graylog
func (m *DockerComposeManager) GELFWriter() (*gelf.Writer, error) {
	return gelf.NewWriter("localhost:12201")
}

// TestHelper provides common test utilities
type TestHelper struct {
	t        require.TestingT
	manager  *DockerComposeManager
	gelfConn *gelf.Writer
}

// NewTestHelper creates a new test helper
func NewTestHelper(t require.TestingT) *TestHelper {
	return &TestHelper{
		t:       t,
		manager: NewDockerComposeManager(context.Background()),
	}
}

// StartGraylog starts Graylog for integration tests
func (h *TestHelper) StartGraylog() {
	if err := h.manager.Start(); err != nil {
		t, ok := h.t.(*testing.T)
		if ok {
			t.Fatalf("Failed to start Graylog: %v", err)
		}
		panic(fmt.Sprintf("Failed to start Graylog: %v", err))
	}
}

// GELFWriter returns a GELF writer connected to Graylog
func (h *TestHelper) GELFWriter() *gelf.Writer {
	if h.gelfConn == nil {
		conn, err := h.manager.GELFWriter()
		if err != nil {
			t, ok := h.t.(*testing.T)
			if ok {
				t.Fatalf("Failed to create GELF writer: %v", err)
			}
			panic(fmt.Sprintf("Failed to create GELF writer: %v", err))
		}
		h.gelfConn = conn
	}
	return h.gelfConn
}

// AssertContainsString checks if expected contains actual (for logging assertions)
func AssertContainsString(t require.TestingT, expected, actual string) {
	if !strings.Contains(actual, expected) {
		t, ok := t.(*testing.T)
		if ok {
			t.Errorf("Expected string to contain %q, got %q", expected, actual)
		}
	}
}

// GenerateTestTraceID generates a deterministic trace ID for tests
func GenerateTestTraceID() string {
	return "test-trace-id-" + fmt.Sprintf("%d", time.Now().UnixNano())
}

// Sleep is a helper for sleeping in tests
func Sleep(d time.Duration) {
	time.Sleep(d)
}


