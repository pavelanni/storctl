package serverchecker

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/types"
)

func TestNewServerChecker(t *testing.T) {
	logger := logger.NewLogger(slog.LevelDebug)
	// Test creation with invalid key path
	_, err := NewServerChecker("localhost:22", config.DefaultAdminUser, "/nonexistent/key", logger, 1*time.Minute, 1)
	if err == nil {
		t.Error("Expected error for nonexistent key, got nil")
	}

	// Test creation with valid parameters (you'll need to provide a real test key)
	// TODO: Add path to a test SSH key
	checker, err := NewServerChecker("localhost:22", "testuser", "testdata/test_key", logger, 1*time.Minute, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if checker == nil {
		t.Error("Expected non-nil ServerChecker")
	}
}

func TestCheckServers(t *testing.T) {
	t.Parallel()

	logger := logger.NewLogger(slog.LevelDebug)
	// Create test servers with mock IPs
	servers := []*types.Server{
		{
			ObjectMeta: types.ObjectMeta{
				Name:   "test-cp",
				Labels: map[string]string{"lab_name": "test"},
			},
			Status: types.ServerStatus{
				PublicNet: &types.PublicNet{
					IPv4: &struct {
						IP string `json:"ip"`
					}{
						IP: "192.0.2.1", // Use TEST-NET-1 range (RFC 5737)
					},
				},
			},
		},
		{
			ObjectMeta: types.ObjectMeta{
				Name:   "test-node-01",
				Labels: map[string]string{"lab_name": "test"},
			},
			Status: types.ServerStatus{
				PublicNet: &types.PublicNet{
					IPv4: &struct {
						IP string `json:"ip"`
					}{
						IP: "192.0.2.2", // Use TEST-NET-1 range (RFC 5737)
					},
				},
			},
		},
	}

	// Use shorter timeout for tests
	results, err := CheckServers(servers, logger, 100*time.Millisecond, 2)
	t.Logf("results: %+v, err: %+v", results, err)

	// Expect error because these are not real servers
	if err == nil {
		t.Error("Expected error for unreachable servers, got nil")
	}

	if len(results) != len(servers) {
		t.Errorf("Expected %d results, got %d", len(servers), len(results))
	}

	// Check that results contain expected errors
	for _, result := range results {
		if result.Ready {
			t.Errorf("Server %s should not be ready", result.Server.ObjectMeta.Name)
		}
		if result.Error == nil {
			t.Errorf("Expected error for server %s", result.Server.ObjectMeta.Name)
		}
	}
}

// Add a new test specifically for checkServerReady
func TestServerChecker_checkServerReady(t *testing.T) {
	t.Parallel()
	logger := logger.NewLogger(slog.LevelDebug)
	// Create a ServerChecker with shorter intervals for testing
	sc, err := NewServerChecker(
		"192.0.2.1:22", // non-routable IP
		"testuser",
		"testdata/test_key",
		logger,
		5*time.Second, // total timeout
		3,             // number of attempts
	)
	if err != nil {
		t.Fatalf("Failed to create ServerChecker: %v", err)
	}

	// Override the ticker duration for testing
	// This is a bit hacky but helps us see more log messages
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a channel to track test completion
	done := make(chan struct{})
	go func() {
		err = sc.checkServerReady(ctx)
		close(done)
	}()

	// Wait for either test completion or timeout
	select {
	case <-done:
		if err == nil {
			t.Error("Expected error for non-routable IP")
		}
	case <-time.After(6 * time.Second):
		t.Error("Test took too long to complete")
	}
}

func TestServerChecker(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses map[string]struct {
			output string
			err    error
		}
		expectReady bool
		expectErr   bool
	}{
		{
			name: "server ready",
			mockResponses: map[string]struct {
				output string
				err    error
			}{
				"cloud-init status --wait": {
					output: "status: done",
					err:    nil,
				},
				"TZ=UTC uptime -s": {
					output: time.Now().UTC().Format("2006-01-02 15:04:05"),
					err:    nil,
				},
				"[ -f /var/run/reboot-required ] && echo 'yes' || echo 'no'": {
					output: "no",
					err:    nil,
				},
				"apt-get -s upgrade | grep -q '^0 upgraded' && echo 'ok' || echo 'pending'": {
					output: "ok",
					err:    nil,
				},
			},
			expectReady: true,
			expectErr:   false,
		},
		{
			name: "cloud-init not done",
			mockResponses: map[string]struct {
				output string
				err    error
			}{
				"cloud-init status --wait": {
					output: "status: running",
					err:    nil,
				},
			},
			expectReady: false,
			expectErr:   true,
		},
		{
			name: "ssh connection failed",
			mockResponses: map[string]struct {
				output string
				err    error
			}{
				"": {
					output: "",
					err:    errors.New("connection refused"),
				},
			},
			expectReady: false,
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSSHClient{
				ConnectFunc: func() error {
					if resp, ok := tt.mockResponses[""]; ok {
						return resp.err
					}
					return nil
				},
				ExecCommandFunc: func(cmd string) (string, error) {
					if resp, ok := tt.mockResponses[cmd]; ok {
						t.Helper()
						t.Log("Command:", cmd)
						t.Log("Output:", resp.output)
						t.Log("Error:", resp.err)
						return resp.output, resp.err
					}
					return "", errors.New("unexpected command")
				},
			}

			// Create logger with debug level
			testLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			checker := &ServerChecker{
				client:         mock,
				host:           "test-host:22",
				attempts:       1,
				timeout:        2000 * time.Millisecond,
				logger:         testLogger,
				tickerDuration: 500 * time.Millisecond,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
			defer cancel()

			ready, err := checker.CheckWithContext(ctx)
			if (err != nil) != tt.expectErr {
				t.Errorf("Check() error = %v, expectErr %v", err, tt.expectErr)
			}
			if ready != tt.expectReady {
				t.Errorf("Check() ready = %v, want %v", ready, tt.expectReady)
			}
		})
	}
}
