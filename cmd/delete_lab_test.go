package cmd

import (
	"testing"

	"github.com/pavelanni/storctl/internal/lab/mock"
	"github.com/pavelanni/storctl/internal/types"
)

func TestDeleteLabCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		flags       []string
		mockSetup   func(*mock.Manager)
		wantErr     bool
		errContains string
	}{
		{
			name:  "successful lab deletion",
			args:  []string{"test-lab"},
			flags: []string{"--yes"}, // Skip confirmation
			mockSetup: func(m *mock.Manager) {
				m.DeleteFunc = func(name string, force bool) error {
					if name != "test-lab" {
						t.Errorf("expected lab name 'test-lab', got '%s'", name)
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:  "force deletion",
			args:  []string{"test-lab"},
			flags: []string{"--yes", "--force"},
			mockSetup: func(m *mock.Manager) {
				m.DeleteFunc = func(name string, force bool) error {
					if !force {
						t.Error("expected force flag to be true")
					}
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:  "lab not found",
			args:  []string{"nonexistent-lab"},
			flags: []string{"--yes"},
			mockSetup: func(m *mock.Manager) {
				m.DeleteFunc = func(name string, force bool) error {
					return types.NewError("NOT_FOUND", "lab not found")
				}
			},
			wantErr:     true,
			errContains: "lab not found",
		},
		{
			name:  "missing lab name",
			args:  []string{},
			flags: []string{"--yes"},
			mockSetup: func(m *mock.Manager) {
				m.DeleteFunc = func(name string, force bool) error {
					t.Error("DeleteLab should not be called when lab name is missing")
					return nil
				}
			},
			wantErr:     true,
			errContains: "requires exactly 1 arg",
		},
		{
			name:  "provider error",
			args:  []string{"test-lab"},
			flags: []string{"--yes"},
			mockSetup: func(m *mock.Manager) {
				m.DeleteFunc = func(name string, force bool) error {
					return types.NewError("PROVIDER_ERROR", "failed to delete lab")
				}
			},
			wantErr:     true,
			errContains: "failed to delete lab",
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
			originalManager := labManager
			labManager = mockManager
			defer func() { labManager = originalManager }()

			// Create and execute the command
			cmd := NewDeleteLabCmd()
			cmd.SetArgs(append(tt.args, tt.flags...))
			err := cmd.Execute()

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteLabCmd() error = %v, wantErr %v", err, tt.wantErr)
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
