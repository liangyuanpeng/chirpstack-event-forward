package integration

import "context"

type Integration interface {
	HandleEvent(ctx context.Context, ch chan HandleError, vars map[string]string, data []byte) (string, error)
	Close() error
}

type HandleError struct {
	Err  error
	Name string
}
