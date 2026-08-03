package platform

import "context"

type ProcessStopper interface {
	StopForSelection(ctx context.Context, platform string, runtimePaths ...string) error
	StopRuntime(ctx context.Context, platform string, runtimePaths ...string) error
}

func NewProcessStopper() ProcessStopper {
	return newProcessStopper()
}
