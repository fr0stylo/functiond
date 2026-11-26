package orchestration

import (
	"context"
	"functiond/pkg/orchestration/pool"
	"log"
	"time"
)

type Worker interface {
	Name() string
	Execute(ctx context.Context, payload []byte) (*ExecResult, error)
	Stop(ctx context.Context) error
}

type WorkerFactory pool.PoolFactory[Worker]

type ExecResult struct {
	Err    error
	Result []byte
}

type WorkerSet struct {
	name string
	ctx  context.Context
	opts *WorkerSetOptions
	pool *pool.Pool[Worker]
}

func (r *WorkerSet) Name() string {
	return r.name
}

func (r *WorkerSet) Start() error {
	return nil
}

func (r *WorkerSet) Execute(ctx context.Context, payload []byte) (*ExecResult, error) {
	l, err := r.pool.Aquire(ctx)
	if err != nil {
		return nil, err
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, r.opts.execTimeout)
	defer cancel()

	t := time.Now()
	b, err := l.Value()
	if err != nil {
		return &ExecResult{Err: err}, err
	}
	key, _ := l.Key()
	resp, err := b.Execute(timeoutCtx, payload)
	if err != nil {
		log.Printf("[%s] timeout period elapsed %s", key, time.Since(t))
		go l.Destroy(ctx)
		return &ExecResult{Err: err}, err
	}

	log.Printf("[%s][%s] %s", key, time.Since(t), resp)
	defer l.Release()

	return resp, nil
}

func (r *WorkerSet) Shutdown(ctx context.Context) error {
	return r.pool.Close(ctx)
}

func NewWorkerSet(ctx context.Context, opts ...Opts[WorkerSetOptions]) (*WorkerSet, error) {
	options := &WorkerSetOptions{}
	DefaultOptions(options)

	for _, o := range opts {
		o(options)
	}

	return &WorkerSet{
		ctx:  ctx,
		opts: options,
		pool: pool.NewPool(
			options.concurrency,
			options.workerFactory,
			options.downscaleTimeout,
		),
	}, nil
}
