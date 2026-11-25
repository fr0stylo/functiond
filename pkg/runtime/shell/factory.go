package shell

import (
	"context"
	"functiond/pkg/orchestration"
	"sync/atomic"
)

// Option configures a shell worker factory.
type Option func(*Factory)

// Factory creates shell workers suitable for orchestration pools.
type Factory struct {
	shell string
	dir   string
	env   []string
	next  atomic.Uint64
}

// NewFactory creates a factory that uses /bin/sh by default.
func NewFactory(opts ...Option) *Factory {
	factory := &Factory{shell: "/bin/sh"}
	for _, opt := range opts {
		opt(factory)
	}
	return factory
}

// WithShell sets the executable used to run each payload.
func WithShell(path string) Option {
	return func(factory *Factory) {
		factory.shell = path
	}
}

// WithDir sets the working directory for shell commands.
func WithDir(path string) Option {
	return func(factory *Factory) {
		factory.dir = path
	}
}

// WithEnv replaces the environment used by shell commands.
func WithEnv(env ...string) Option {
	return func(factory *Factory) {
		factory.env = append([]string(nil), env...)
	}
}

// Create constructs and starts a worker. It implements pool.PoolFactory.
func (factory *Factory) Create(ctx context.Context, name string) (orchestration.Worker, error) {
	worker := &Shell{
		name:  name,
		shell: factory.shell,
		dir:   factory.dir,
		env:   append([]string(nil), factory.env...),
	}
	if err := worker.Init(ctx); err != nil {
		return nil, err
	}
	if err := worker.Start(ctx); err != nil {
		return nil, err
	}
	return worker, nil
}
