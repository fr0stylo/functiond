package runner

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"

	"functiond/pkg/runner/worker"
)

type WorkerSetOptions struct {
	name        string
	initCommand string
	filePath    string
	concurrency int
	ports       []worker.PortMapping
}

var defaultWSOptions = WorkerSetOptions{
	name:        "nodejs",
	filePath:    "",
	concurrency: 5,
}

func WithWorkerSetName(name string) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		opts.name = name
	}
}

func WithFile(filePath string) worker.Opts[WorkerSetOptions] {
	return func(opts *WorkerSetOptions) {
		opts.filePath = filePath
	}
}

type WorkerSet struct {
	WorkerSetOptions
	ctx            context.Context
	client         *containerd.Client
	networkManager *worker.NetworkManager
	workers        chan worker.Worker
	snapshotName   string
	workerCount    int
	lock           *sync.Mutex
}

func (r *WorkerSet) Start() error {
	node, _ := worker.NewNode(r.client, r.networkManager, worker.WithLabels(map[string]string{
		"RUNNING":       "example",
		"IgnoreUnknown": "1",
	}), worker.WithName(r.name+"-"+randSeq(4)), worker.WithSnapshot(r.snapshotName))
	r.workerCount++
	r.workers <- node
	return node.Start(r.ctx)
}

func (r *WorkerSet) Execute(ctx context.Context, payload []byte) error {
	if r.workerCount < r.WorkerSetOptions.concurrency && len(r.workers) == 0 {
		if err := r.Start(); err != nil {
			return err
		}
	}
	w := <-r.workers
	go func() {
		t := time.Now()
		b, _ := w.Execute(ctx, payload)
		log.Printf("[%s][%s] %s", w.Name(), time.Since(t), b)
		defer func() { r.workers <- w }()
	}()

	return nil
}

func (r *WorkerSet) Shutdown() {
	defer r.client.Close()
	close(r.workers)

	for w := range r.workers {
		w.Shutdown(r.ctx)
	}
}

func NewWorkerSet(opts ...worker.Opts[WorkerSetOptions]) (*WorkerSet, error) {
	options := defaultWSOptions
	for _, optFn := range opts {
		optFn(&options)
	}

	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		return nil, err
	}
	nm, err := worker.NewNetworkManager()
	if err != nil {
		return nil, err
	}
	ctx := namespaces.WithNamespace(context.Background(), "example")
	snapshot, err := NewSnapshot(client).
		CreateSnapshot(ctx,
			"docker.io/library/node:lts-alpine",
			options.name,
			"final",
			options.filePath)

	return &WorkerSet{
		ctx:              ctx,
		client:           client,
		networkManager:   nm,
		WorkerSetOptions: options,
		snapshotName:     snapshot,
		workers:          make(chan worker.Worker, options.concurrency),
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
