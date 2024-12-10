package serverchecker

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type RealSSHClient struct {
	host    string
	user    string
	keyPath string
	client  *ssh.Client
}

func (r *RealSSHClient) Connect() error {
	key, err := os.ReadFile(r.keyPath)
	if err != nil {
		return fmt.Errorf("reading SSH key %s: %w", r.keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("parsing SSH key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: r.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", r.host, config)
	if err != nil {
		return fmt.Errorf("SSH dial: %w", err)
	}
	r.client = client
	return nil
}

func (r *RealSSHClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *RealSSHClient) ExecCommand(cmd string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("client not connected")
	}

	session, err := r.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	defer session.Close()

	var output bytes.Buffer
	session.Stdout = &output
	if err := session.Run(cmd); err != nil {
		return output.String(), fmt.Errorf("run command: %w", err)
	}
	return output.String(), nil
}
