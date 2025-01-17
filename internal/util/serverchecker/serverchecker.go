package serverchecker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
	"golang.org/x/crypto/ssh"
)

type ServerChecker struct {
	client         SSHClient
	host           string
	attempts       int
	timeout        time.Duration
	logger         *slog.Logger
	tickerDuration time.Duration
}

type ServerResult struct {
	Server *types.Server
	Ready  bool
	Error  error
}

func NewServerChecker(host string, user string, keyPath string, logger *slog.Logger, timeout time.Duration, attempts int) (*ServerChecker, error) {
	// check if key exists
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("key file does not exist: %s", keyPath)
	}

	// Create real SSH client here
	client := &RealSSHClient{
		host:    host,
		user:    user,
		keyPath: keyPath,
	}

	return &ServerChecker{
		client:   client,
		host:     host,
		attempts: attempts,
		timeout:  timeout,
		logger:   logger,
	}, nil
}

func CheckServers(servers []*types.Server, logger *slog.Logger, timeout time.Duration, attempts int) ([]ServerResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make([]ServerResult, len(servers))
	for i, server := range servers {
		wg.Add(1)
		serverIP := server.Status.PublicNet.IPv4.IP
		serverPrivateKeyPath := filepath.Join(os.Getenv("HOME"),
			config.DefaultConfigDir,
			config.DefaultKeysDir,
			strings.Join([]string{server.ObjectMeta.Labels["lab_name"], "admin"}, "-"))
		if serverIP == "" {
			results[i] = ServerResult{Server: server, Error: fmt.Errorf("server IP is empty")}
			continue
		}

		go func(i int, server *types.Server) {
			defer wg.Done()
			sc, err := NewServerChecker(serverIP+":22", config.DefaultAdminUser, serverPrivateKeyPath, logger, timeout, attempts)
			if err != nil {
				results[i] = ServerResult{Server: server, Error: err}
				return
			}
			err = sc.checkServerReady(ctx)
			if err != nil {
				results[i] = ServerResult{Server: server, Error: err}
				return
			}
			results[i] = ServerResult{Server: server, Ready: true}
		}(i, server)
	}
	wg.Wait()
	for _, result := range results {
		if result.Error != nil {
			return results, result.Error
		}
	}
	return results, nil
}

func (sc *ServerChecker) checkServerReady(ctx context.Context) error {
	tickerDuration := 30 * time.Second
	if sc.tickerDuration > 0 {
		tickerDuration = sc.tickerDuration
	}
	ticker := time.NewTicker(tickerDuration)
	defer ticker.Stop()

	timeout := time.After(sc.timeout)
	attempt := 0

	sc.logger.Info("Starting server readiness check",
		"host", sc.host,
		"timeout", sc.timeout,
		"maxAttempts", sc.attempts)

	for {
		select {
		case <-ctx.Done():
			sc.logger.Debug("Check cancelled by context",
				"host", sc.host,
				"error", ctx.Err())
			return ctx.Err()
		case <-timeout:
			sc.logger.Error("Timeout reached",
				"host", sc.host,
				"timeout", sc.timeout,
				"attemptsMade", attempt)
			return fmt.Errorf("timeout waiting for server to be ready after %d attempts", attempt)
		case <-ticker.C:
			attempt++
			sc.logger.Debug("Starting connection attempt",
				"host", sc.host,
				"attempt", attempt,
				"maxAttempts", sc.attempts)

			if attempt > sc.attempts {
				sc.logger.Error("Max attempts reached",
					"host", sc.host,
					"attempts", sc.attempts)
				return fmt.Errorf("max attempts (%d) reached", sc.attempts)
			}

			err := sc.client.Connect()
			if err != nil {
				sc.logger.Debug("SSH connection failed",
					"host", sc.host,
					"attempt", attempt,
					"error", err,
					"timeElapsed", time.Since(time.Now()))
				continue
			}
			defer sc.client.Close()

			// Check cloud-init status
			cloudInitStatus, err := sc.client.ExecCommand("cloud-init status --wait")
			if err != nil {
				if exitErr, ok := err.(*ssh.ExitError); ok && exitErr.ExitStatus() == 2 {
					// Exit code 2 means warning, which we'll accept
					sc.logger.Debug("Cloud-init completed with warnings", "host", sc.host, "attempt", attempt)
				} else {
					sc.logger.Debug("Cloud-init check failed", "host", sc.host, "attempt", attempt, "error", err)
					continue
				}
			}
			if !strings.Contains(cloudInitStatus, "done") {
				sc.logger.Debug("Cloud-init not done yet", "host", sc.host, "attempt", attempt, "status", strings.TrimSpace(cloudInitStatus))
				continue
			}

			// Check boot time
			uptime, err := sc.client.ExecCommand("TZ=UTC uptime -s")
			if err != nil {
				sc.logger.Debug("Uptime check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}

			bootTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(uptime))
			if err != nil {
				sc.logger.Debug("Failed to parse boot time", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}

			sc.logger.Debug("Boot time", "host", sc.host, "attempt", attempt, "bootTime", bootTime)
			if time.Since(bootTime) > 5*time.Minute {
				sc.logger.Debug("Server boot time too old", "host", sc.host, "attempt", attempt, "timeSinceBoot", time.Since(bootTime))
				continue
			}

			// Check for pending updates
			rebootRequired, err := sc.client.ExecCommand("[ -f /var/run/reboot-required ] && echo 'yes' || echo 'no'")
			if err != nil {
				sc.logger.Debug("Reboot check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}
			if strings.TrimSpace(rebootRequired) == "yes" {
				sc.logger.Debug("Reboot still required", "host", sc.host, "attempt", attempt)
				continue
			}

			// Check package status
			pkgStatus, err := sc.client.ExecCommand("apt-get -s upgrade | grep -q '^0 upgraded' && echo 'ok' || echo 'pending'")
			if err != nil {
				sc.logger.Debug("Package status check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}
			if strings.TrimSpace(pkgStatus) != "ok" {
				sc.logger.Debug("Packages still need upgrading", "host", sc.host, "attempt", attempt)
				continue
			}

			sc.logger.Info("Server is ready!", "host", sc.host, "bootTime", bootTime)
			return nil
		}
	}
}

// CheckWithContext verifies if the server is ready using the provided context
// Returns true if the server is ready, false otherwise
// If an error occurs during checking, it returns false and the error
func (sc *ServerChecker) CheckWithContext(ctx context.Context) (bool, error) {
	// Ensure at least one attempt
	err := sc.checkServerReady(ctx)
	if err == nil {
		return true, nil
	}

	// Continue with remaining attempts if time permits
	attempts := 1
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for attempts < sc.attempts {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("timeout waiting for server to be ready after %d attempts", attempts)
		case <-ticker.C:
			err := sc.checkServerReady(ctx)
			if err == nil {
				return true, nil
			}
			attempts++
			sc.logger.Debug("Server not ready yet", "host", sc.host, "attempt", attempts)
		}
	}

	return false, fmt.Errorf("server not ready after %d attempts", attempts)
}

// Check verifies if the server is ready
// This is now a wrapper around CheckWithContext using the server's timeout
func (sc *ServerChecker) Check() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sc.timeout)
	defer cancel()
	return sc.CheckWithContext(ctx)
}
