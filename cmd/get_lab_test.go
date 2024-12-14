package cmd

import (
	"testing"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/lab/mock"
	"github.com/pavelanni/storctl/internal/types"
)

func TestGetLabCmd(t *testing.T) {
	// Create some test data
	testTime := time.Now()
	testLabs := []*types.Lab{
		{
			ObjectMeta: types.ObjectMeta{
				Name: "test-lab-1",
				Labels: map[string]string{
					"environment": "staging",
				},
			},
			Status: types.LabStatus{
				Created: testTime,
				State:   "running",
				Servers: []*types.Server{
					{
						ObjectMeta: types.ObjectMeta{
							Name: "server-1",
						},
						Spec: types.ServerSpec{
							ServerType: "cx11",
						},
						Status: types.ServerStatus{
							Cores:  2,
							Memory: 4,
							Disk:   50,
						},
					},
				},
				Volumes: []*types.Volume{
					{
						ObjectMeta: types.ObjectMeta{
							Name: "volume-1",
						},
						Spec: types.VolumeSpec{
							Size: 100,
						},
					},
				},
			},
		},
		{
			ObjectMeta: types.ObjectMeta{
				Name: "test-lab-2",
				Labels: map[string]string{
					"environment": "production",
				},
			},
			Status: types.LabStatus{
				Created: testTime,
				State:   "stopped",
			},
		},
	}

	// Save original config and restore after tests
	originalCfg := cfg
	defer func() { cfg = originalCfg }()

	// Initialize config for tests
	cfg = &config.Config{
		OutputFormat: "table",
	}

	tests := []struct {
		name         string
		args         []string
		outputFormat string
		mockSetup    func(*mock.Manager)
		wantErr      bool
		errContains  string
	}{
		{
			name: "list all labs",
			args: []string{},
			mockSetup: func(m *mock.Manager) {
				m.ListFunc = func() ([]*types.Lab, error) {
					return testLabs, nil
				}
			},
			wantErr: false,
		},
		{
			name:         "get specific lab",
			args:         []string{"test-lab-1"},
			outputFormat: "table",
			mockSetup: func(m *mock.Manager) {
				m.GetFunc = func(name string) (*types.Lab, error) {
					if name == "test-lab-1" {
						return testLabs[0], nil
					}
					return nil, types.NewError("NOT_FOUND", "lab not found")
				}
			},
			wantErr: false,
		},
		{
			name:         "get lab in JSON format",
			args:         []string{"test-lab-1"},
			outputFormat: "json",
			mockSetup: func(m *mock.Manager) {
				m.GetFunc = func(name string) (*types.Lab, error) {
					return testLabs[0], nil
				}
			},
			wantErr: false,
		},
		{
			name:         "get lab in YAML format",
			args:         []string{"test-lab-1"},
			outputFormat: "yaml",
			mockSetup: func(m *mock.Manager) {
				m.GetFunc = func(name string) (*types.Lab, error) {
					return testLabs[0], nil
				}
			},
			wantErr: false,
		},
		{
			name: "lab not found",
			args: []string{"nonexistent-lab"},
			mockSetup: func(m *mock.Manager) {
				m.GetFunc = func(name string) (*types.Lab, error) {
					return nil, types.NewError("NOT_FOUND", "lab not found")
				}
			},
			wantErr:     true,
			errContains: "lab not found",
		},
		{
			name: "provider error during list",
			args: []string{},
			mockSetup: func(m *mock.Manager) {
				m.ListFunc = func() ([]*types.Lab, error) {
					return nil, types.NewError("PROVIDER_ERROR", "failed to list labs")
				}
			},
			wantErr:     true,
			errContains: "failed to list labs",
		},
		{
			name: "get lab from cloud",
			args: []string{"test-lab-1", "--from-cloud"},
			mockSetup: func(m *mock.Manager) {
				m.GetFromCloudFunc = func(name string) (*types.Lab, error) {
					if name == "test-lab-1" {
						return testLabs[0], nil
					}
					return nil, types.NewError("NOT_FOUND", "lab not found in cloud")
				}
			},
			wantErr: false,
		},
		{
			name: "get lab from cloud - not found",
			args: []string{"nonexistent-lab", "--from-cloud"},
			mockSetup: func(m *mock.Manager) {
				m.GetFromCloudFunc = func(name string) (*types.Lab, error) {
					return nil, types.NewError("NOT_FOUND", "lab not found in cloud")
				}
			},
			wantErr:     true,
			errContains: "lab not found in cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock lab manager for each test
			mockManager := &mock.Manager{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockManager)
			}

			// Set output format for this test
			cfg.OutputFormat = tt.outputFormat

			// Store the current lab manager and restore it after the test
			// TODO: fix it later
			//originalManager := lab.DefaultManager
			//lab.DefaultManager = &lab.ManagerSvc{
			//	Provider: mockManager,
			//}
			//defer func() { lab.DefaultManager = originalManager }()

			// Create and execute the command
			cmd := NewGetLabCmd()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLabCmd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
			}
		})
	}
}
