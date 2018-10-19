package providers

import (
	"fmt"

	"github.com/pkg/errors"
)

type ProviderError struct {
	ParameterErrors []error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("provider error: %v", e.ParameterErrors)
}

func NewParameterError(parameterName string, reason error) error {
	return errors.Errorf("failed loading key '%s': %v", parameterName, reason)
}
