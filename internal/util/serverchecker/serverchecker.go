package serverchecker

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/logger"
	"github.com/pavelanni/labshop/internal/types"
	"golang.org/x/crypto/ssh"
)

type ServerChecker struct {
	host     string
	config   *ssh.ClientConfig
	timeout  time.Duration
	attempts int
	logger   *slog.Logger
}

type ServerResult struct {
	Server *types.Server
	Ready  bool
	Error  error
}

func NewServerChecker(host, user, keyPath, logLevel string, timeout time.Duration, attempts int) (*ServerChecker, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key %s: %w", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	level := logger.ParseLevel(logLevel)
	serverCheckerLogger := logger.NewLogger(level)
	return &ServerChecker{
		host:     host,
		config:   config,
		timeout:  timeout,
		attempts: attempts,
		logger:   serverCheckerLogger,
	}, nil
}

func CheckServers(servers []*types.Server, logLevel string, timeout time.Duration, attempts int) ([]ServerResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup
	results := make([]ServerResult, len(servers))
	for i, server := range servers {
		wg.Add(1)
		serverIP := server.Status.PublicNet.IPv4.IP
		serverPrivateKeyPath := filepath.Join(os.Getenv("HOME"),
			config.DefaultConfigDir,
			config.KeysDir,
			strings.Join([]string{server.ObjectMeta.Labels["lab_name"], "admin"}, "-"))
		go func(i int, server *types.Server) {
			defer wg.Done()
			sc, err := NewServerChecker(serverIP+":22", config.DefaultAdminUser, serverPrivateKeyPath, logLevel, 30*time.Minute, 20)
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
	return results, nil
}

func (sc *ServerChecker) execCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output
	if err := session.Run(cmd); err != nil {
		return output.String(), err
	}
	return output.String(), nil
}

func (sc *ServerChecker) checkServerReady(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
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

			client, err := ssh.Dial("tcp", sc.host, sc.config)
			if err != nil {
				sc.logger.Debug("SSH connection failed",
					"host", sc.host,
					"attempt", attempt,
					"error", err,
					"timeElapsed", time.Since(time.Now()))
				continue
			}
			defer client.Close()

			// Check cloud-init status
			cloudInitStatus, err := sc.execCommand(client, "cloud-init status --wait")
			if err != nil {
				sc.logger.Debug("Cloud-init check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}
			if !strings.Contains(cloudInitStatus, "done") {
				sc.logger.Debug("Cloud-init not done yet", "host", sc.host, "attempt", attempt, "status", strings.TrimSpace(cloudInitStatus))
				continue
			}

			// Check boot time
			uptime, err := sc.execCommand(client, "uptime -s")
			if err != nil {
				sc.logger.Debug("Uptime check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}

			bootTime, err := time.Parse("2006-01-02 15:04:05", strings.TrimSpace(uptime))
			if err != nil {
				sc.logger.Debug("Failed to parse boot time", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}

			if time.Since(bootTime) > 5*time.Minute {
				sc.logger.Debug("Server boot time too old", "host", sc.host, "attempt", attempt)
				continue
			}

			// Check for pending updates
			rebootRequired, err := sc.execCommand(client, "[ -f /var/run/reboot-required ] && echo 'yes' || echo 'no'")
			if err != nil {
				sc.logger.Debug("Reboot check failed", "host", sc.host, "attempt", attempt, "error", err)
				continue
			}
			if strings.TrimSpace(rebootRequired) == "yes" {
				sc.logger.Debug("Reboot still required", "host", sc.host, "attempt", attempt)
				continue
			}

			// Check package status
			pkgStatus, err := sc.execCommand(client, "apt-get -s upgrade | grep -q '^0 upgraded' && echo 'ok' || echo 'pending'")
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
