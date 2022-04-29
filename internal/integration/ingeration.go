package integration

import "context"

type Integration interface {
	HandleEvent(ctx context.Context, vars map[string]string, data []byte) (string, error)
	Close() error
}
