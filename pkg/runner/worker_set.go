package runner

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	containerd "github.com/containerd/containerd/v2/client"

	"functiond/pkg/runner/snapshotter"
	"functiond/pkg/runner/worker"
)

type WorkerSetOptions struct {
	runtime              string
	name                 string
	environmentVariables map[string]string
	filePath             string
	initCommand          []string
	execTimeout          int
	memoryLimit          int
	cpuLimit             int
	concurrency          int
	downscaleTimeout     time.Duration
}

var defaultWSOptions = WorkerSetOptions{
	name:                 "nodejs",
	initCommand:          []string{"node", "lambda.js"},
	environmentVariables: map[string]string{},
	execTimeout:          5,
	filePath:             "",
	concurrency:          110,
	downscaleTimeout:     10 * time.Second,
}

func WithWorkerSetName(name string) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		opts.name = name
	}
}

func WithOptions(options *WorkerSetOptions) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		*opts = *options
	}
}

func WithFile(filePath string) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		opts.filePath = filePath
	}
}

func WithDownscaleTimeout(duration time.Duration) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		opts.downscaleTimeout = duration
	}
}

func BuildOptions(opts ...worker.Opts[WorkerSetOptions]) *WorkerSetOptions {
	options := defaultWSOptions
	for _, optFn := range opts {
		optFn(&options)
	}

	return &options
}

type WorkerSet struct {
	*WorkerSetOptions
	ctx            context.Context
	client         *containerd.Client
	networkManager *worker.NetworkManager
	workers        chan worker.Worker
	snapshotName   string
	downscaleMap   map[string]*time.Timer
	lock           sync.Mutex
}

func (r *WorkerSet) Name() string {
	return r.name
}

func (r *WorkerSet) killWorker(node worker.Worker) {
	r.downscaleMap[node.Name()].Stop()
	delete(r.downscaleMap, node.Name())
	node.Shutdown(r.ctx)
	for {
		w := <-r.workers
		if w.Name() == node.Name() {
			return
		}
		r.workers <- w
	}
}

func (r *WorkerSet) Start() error {
	node, _ := worker.NewNode(r.client, r.networkManager, worker.WithLabels(map[string]string{
		"RUNNING":       "example",
		"IgnoreUnknown": "1",
	}), worker.WithName(r.name+"-"+randSeq(4)), worker.WithSnapshot(r.snapshotName))
	if err := node.Start(r.ctx); err != nil {
		return err
	}
	r.workers <- node
	r.downscaleMap[node.Name()] = time.AfterFunc(r.downscaleTimeout, func() {
		r.killWorker(node)
	})
	return nil
}

type ExecResult struct {
	Err    error
	Result []byte
}

func (r *WorkerSet) Execute(ctx context.Context, payload []byte) (chan *ExecResult, error) {
	res := make(chan *ExecResult)
	w, err := r.retrieveWorker()
	if err != nil {
		return nil, err
	}
	go func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(r.execTimeout)*time.Second)
		defer cancel()

		t := time.Now()
		b, err := w.Execute(timeoutCtx, payload)
		if err != nil {
			log.Printf("[%s] timeout period elapsed %s", w.Name(), time.Since(t))
			res <- &ExecResult{Err: err}
			go r.killWorker(w)
			return
		}

		log.Printf("[%s][%s] %s", w.Name(), time.Since(t), b)
		defer func() {
			r.downscaleMap[w.Name()].Reset(r.downscaleTimeout)
			r.workers <- w
		}()
		res <- &ExecResult{Result: b}
	}()

	return res, nil
}

func (r *WorkerSet) retrieveWorker() (worker.Worker, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if len(r.downscaleMap) < r.WorkerSetOptions.concurrency && len(r.workers) == 0 {
		if err := r.Start(); err != nil {
			return nil, err
		}

	}
	w := <-r.workers
	if r.downscaleMap[w.Name()] == nil {
		return r.retrieveWorker()
	}

	return w, nil
}

func (r *WorkerSet) Shutdown() {
	defer r.client.Close()
	close(r.workers)

	for w := range r.workers {
		if r.downscaleMap[w.Name()] != nil {
			w.Shutdown(r.ctx)
		}
	}
}

func NewWorkerSet(ctx context.Context, opts ...worker.Opts[WorkerSetOptions]) (*WorkerSet, error) {
	options := BuildOptions(opts...)

	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}
	nm, err := worker.NewNetworkManager()
	if err != nil {
		return nil, err
	}

	snapshot, err := snapshotter.NewSnapshot(client).
		CreateSnapshot(ctx,
			"docker.io/library/node:lts-alpine",
			options.name,
			"final",
			options.filePath)
	if err != nil {
		return nil, err
	}

	return &WorkerSet{
		ctx:              ctx,
		client:           client,
		networkManager:   nm,
		WorkerSetOptions: options,
		snapshotName:     snapshot,
		workers:          make(chan worker.Worker, options.concurrency),
		downscaleMap:     make(map[string]*time.Timer),
	}, nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
