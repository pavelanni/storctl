package cmd

import (
	"strings"
	"testing"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/lab/mock"
	"github.com/pavelanni/storctl/internal/types"
)

func TestCreateLabCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		mockSetup   func(*mock.Manager)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful lab creation",
			args: []string{"test-lab", "--template", "lab.yaml"},
			mockSetup: func(m *mock.Manager) {
				m.CreateFunc = func(lab *types.Lab) error {
					if lab.ObjectMeta.Name != "test-lab" {
						t.Errorf("expected lab name 'test-lab', got '%s'", lab.ObjectMeta.Name)
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "provider error",
			args: []string{"test-lab", "--template", "lab.yaml"},
			mockSetup: func(m *mock.Manager) {
				m.CreateFunc = func(lab *types.Lab) error {
					return types.NewError("provider error", "failed to create lab")
				}
			},
			wantErr:     true,
			errContains: "failed to create lab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock provider for each test
			mockManager := &mock.Manager{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockManager)
			}

			// Store the current provider and restore it after the test
			originalManager := labSvc
			labSvc = mockManager.ManagerSvc
			defer func() { labSvc = originalManager }()

			// Store the current config and restore it after the test
			originalCfg := cfg
			cfg = &config.Config{
				Owner:        "test-owner",
				Organization: "test-organization",
				Email:        "test-email",
			}
			defer func() { cfg = originalCfg }()

			// Create and execute the command
			cmd := NewCreateLabCmd()
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLabCmd() error = %v, wantErr %v", err, tt.wantErr)
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

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
