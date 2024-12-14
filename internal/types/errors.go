package types

// ProviderError represents an error from a cloud provider
type ProviderError struct {
	Code    string
	Message string
}

func (e *ProviderError) Error() string {
	return e.Message
}

// NewError creates a new ProviderError
func NewError(code, message string) error {
	return &ProviderError{
		Code:    code,
		Message: message,
	}
}
