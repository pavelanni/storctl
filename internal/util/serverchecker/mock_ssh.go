package serverchecker

type MockSSHClient struct {
	ConnectFunc     func() error
	CloseFunc       func() error
	ExecCommandFunc func(string) (string, error)
}

func (m *MockSSHClient) Connect() error {
	if m.ConnectFunc != nil {
		return m.ConnectFunc()
	}
	return nil
}

func (m *MockSSHClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockSSHClient) ExecCommand(cmd string) (string, error) {
	if m.ExecCommandFunc != nil {
		return m.ExecCommandFunc(cmd)
	}
	return "", nil
}
