// Package ssh provides functions to manage local SSH keys.
package ssh

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"golang.org/x/crypto/ssh"
)

type Manager struct {
	keysDir string
	logger  *slog.Logger
}

func NewManager(cfg *config.Config) *Manager {
	logLevel := logger.ParseLevel(cfg.LogLevel)
	return &Manager{
		keysDir: filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.DefaultKeysDir),
		logger:  logger.NewLogger(logLevel),
	}
}

// CreateLocalKeyPair creates a local SSH key pair
// Returns the public key string in OpenSSH format.
func (m *Manager) CreateLocalKeyPair(name string) (publicKey string, err error) {
	// Generate the key pair
	pubKey, privKey, err := generateED25519KeyPair(name)
	if err != nil {
		return "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Save the keys locally
	if err := os.MkdirAll(m.keysDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create keys directory: %w", err)
	}

	// Save private key
	privKeyPath := filepath.Join(m.keysDir, name)
	if err := os.WriteFile(privKeyPath, privKey, 0600); err != nil {
		return "", fmt.Errorf("failed to save private key: %w", err)
	}

	m.logger.Debug("created local key pair",
		"name", name,
		"path", privKeyPath)

	return string(pubKey), nil
}

// ReadLocalPublicKey reads a local public SSH key.
func (m *Manager) ReadLocalPublicKey(name string) (string, error) {
	pubKeyPath := filepath.Join(m.keysDir, name+".pub")
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read local public key: %w", err)
	}
	return string(pubKey), nil
}

// DeleteLocalKeyPair deletes a local SSH key pair.
func (m *Manager) DeleteLocalKeyPair(name string) error {
	privKeyPath := filepath.Join(m.keysDir, name)
	pubKeyPath := filepath.Join(m.keysDir, name+".pub")

	m.logger.Debug("deleting local key pair",
		"name", name,
		"private_key", privKeyPath,
		"public_key", pubKeyPath)

	if err := m.deleteKeyFile(privKeyPath); err != nil {
		return err
	}
	if err := m.deleteKeyFile(pubKeyPath); err != nil {
		return err
	}
	return nil
}

// LocalKeyExists checks if a local SSH key pair exists.
func (m *Manager) LocalKeyExists(name string) (bool, error) {
	privKeyPath := filepath.Join(m.keysDir, name)
	pubKeyPath := filepath.Join(m.keysDir, name+".pub")
	_, err := os.Stat(privKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check local private SSH key existence: %w", err)
	}
	_, err = os.Stat(pubKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check local public SSH key existence: %w", err)
	}
	return true, nil
}

func (m *Manager) deleteKeyFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			m.logger.Error("failed to delete key file",
				"path", path,
				"error", err)
			return fmt.Errorf("failed to delete %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check %s existence: %w", path, err)
	}
	return nil
}

// generateED25519KeyPair generates a new ED25519 keypair.
// Returns public key in OpenSSH format and private key in PEM format as byte slices.
func generateED25519KeyPair(comment string) (publicKey, privateKey []byte, err error) {
	// Generate the keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ED25519 keypair: %w", err)
	}

	// Convert to SSH private key format and encode as PEM
	pemBlock, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), comment)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encode private key in PEM format
	privateKey = pem.EncodeToMemory(pemBlock)
	if privateKey == nil {
		return nil, nil, fmt.Errorf("failed to encode private key")
	}

	// Generate the public key
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create public key: %w", err)
	}

	// Format public key in OpenSSH format: "ssh-ed25519 <base64> comment"
	pubKey := fmt.Sprintf("%s %s", sshPub.Type(),
		base64.StdEncoding.EncodeToString(sshPub.Marshal()))
	if comment != "" {
		pubKey = fmt.Sprintf("%s %s", pubKey, comment)
	}

	return []byte(pubKey), privateKey, nil
}
