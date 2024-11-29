package serverchecker

import (
	"context"
	"testing"
	"time"

	"github.com/pavelanni/labshop/internal/types"
)

func TestNewServerChecker(t *testing.T) {
	// Test creation with invalid key path
	_, err := NewServerChecker("localhost:22", "testuser", "/nonexistent/key", 1*time.Minute, 1)
	if err == nil {
		t.Error("Expected error for nonexistent key, got nil")
	}

	// Test creation with valid parameters (you'll need to provide a real test key)
	// TODO: Add path to a test SSH key
	checker, err := NewServerChecker("localhost:22", "testuser", "testdata/test_key", 1*time.Minute, 1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if checker == nil {
		t.Error("Expected non-nil ServerChecker")
	}
}

func TestCheckServers(t *testing.T) {
	t.Parallel()

	// Create test servers with real IPs
	servers := []*types.Server{
		{
			ObjectMeta: types.ObjectMeta{
				Name:   "my-lab-cp",
				Labels: map[string]string{"labname": "my-lab"},
			},
			Status: types.ServerStatus{
				PublicNet: &types.PublicNet{
					IPv4: &struct {
						IP string `json:"ip"`
					}{
						IP: "188.245.99.174",
					},
				},
			},
		},
		{
			ObjectMeta: types.ObjectMeta{
				Name:   "my-lab-node-01",
				Labels: map[string]string{"labname": "my-lab"},
			},
			Status: types.ServerStatus{
				PublicNet: &types.PublicNet{
					IPv4: &struct {
						IP string `json:"ip"`
					}{
						IP: "78.46.138.190",
					},
				},
			},
		},
	}

	// Create a context with a longer timeout for real servers
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run the check
	results, err := CheckServers(ctx, servers)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(results) != len(servers) {
		t.Errorf("Expected %d results, got %d", len(servers), len(results))
	}

	// Check each result
	for i, result := range results {
		t.Logf("Checking server %s (%s)", result.Server.ObjectMeta.Name, result.Server.Status.PublicNet.IPv4.IP)
		if result.Server != servers[i] {
			t.Errorf("Server mismatch for %s", result.Server.ObjectMeta.Name)
		}
		if result.Error != nil {
			t.Errorf("Server %s check failed: %v", result.Server.ObjectMeta.Name, result.Error)
		}
		if !result.Ready {
			t.Errorf("Server %s is not ready", result.Server.ObjectMeta.Name)
		}
	}
}

// Add a new test specifically for checkServerReady
func TestServerChecker_checkServerReady(t *testing.T) {
	t.Parallel()

	// Create a ServerChecker with shorter intervals for testing
	sc, err := NewServerChecker(
		"192.0.2.1:22", // non-routable IP
		"testuser",
		"testdata/test_key",
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
