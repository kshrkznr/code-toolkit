//go:build !darwin && !windows

package platform

import (
	"context"
	"fmt"
)

type unsupportedProcessStopper struct{}

func newProcessStopper() ProcessStopper { return unsupportedProcessStopper{} }

func (unsupportedProcessStopper) StopForSelection(context.Context, string, ...string) error {
	return fmt.Errorf("platform process management is unsupported on this host")
}

func (unsupportedProcessStopper) StopRuntime(context.Context, string, ...string) error {
	return fmt.Errorf("platform process management is unsupported on this host")
}
