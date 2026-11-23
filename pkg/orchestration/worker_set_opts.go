package orchestration

import "time"

type WorkerOpts interface{}

type WorkerSetOptions struct {
	execTimeout      time.Duration
	memoryLimit      int
	cpuLimit         int
	concurrency      int
	downscaleTimeout time.Duration
	workerOpts       WorkerOpts
}

var defaultWSOptions = WorkerSetOptions{
	execTimeout:      5,
	concurrency:      110,
	downscaleTimeout: 10 * time.Second,
}

func buildWorkerSetOptions(opts ...Opts[WorkerSetOptions]) *WorkerSetOptions {
	options := defaultWSOptions
	for _, optFn := range opts {
		optFn(&options)
	}

	return &options
}

func DefaultOptions(opts *WorkerSetOptions) {
	opts.concurrency = 10
	opts.cpuLimit = 1
	opts.memoryLimit = 256
	opts.downscaleTimeout = 1 * time.Minute
	opts.execTimeout = 5 * time.Second
	opts.workerOpts = nil
}

func WithConcurency(i int) func(opts *WorkerSetOptions) {
	return func(opts *WorkerSetOptions) {
		opts.concurrency = i
	}
}

func WithCPULimit(i int) func(opts *WorkerSetOptions) {
	return func(opts *WorkerSetOptions) {
		opts.cpuLimit = i
	}
}

func WithMemLimit(i int) func(opts *WorkerSetOptions) {
	return func(opts *WorkerSetOptions) {
		opts.memoryLimit = i
	}
}

func WithTimeout(i time.Duration) func(opts *WorkerSetOptions) {
	return func(opts *WorkerSetOptions) {
		opts.execTimeout = i
	}
}

func WithDownscaleTimeout(i time.Duration) func(opts *WorkerSetOptions) {
	return func(opts *WorkerSetOptions) {
		opts.downscaleTimeout = i
	}
}
