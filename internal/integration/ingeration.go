package integration

import "context"

type Integration interface {
	HandleEvent(ctx context.Context, data []byte) error
	Close() error
}
