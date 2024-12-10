package serverchecker

// SSHClient interface abstracts SSH operations
type SSHClient interface {
	Connect() error
	Close() error
	ExecCommand(cmd string) (string, error)
}
