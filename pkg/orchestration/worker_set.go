package orchestration

import (
	"context"
	"math/rand"
)

type Worker interface {
	Name() string
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Execute(ctx context.Context, payload []byte) (*ExecResult, error)
	Stop(ctx context.Context) error
}

type ExecResult struct {
	Err    error
	Result []byte
}

type WorkerSet struct {
	name string
	ctx  context.Context
	opts *WorkerSetOptions
}

func (r *WorkerSet) Name() string {
	return r.name
}

func (r *WorkerSet) killWorker(node Worker) {
	// r.downscaleMap[node.Name()].Stop()
	// delete(r.downscaleMap, node.Name())
	// node.Shutdown(r.ctx)
	// for {
	// 	w := <-r.workers
	// 	if w.Name() == node.Name() {
	// 		return
	// 	}
	// 	r.workers <- w
	// }
}

func (r *WorkerSet) Start() error {
	// node, _ := worker.NewNode(r.client, r.networkManager, worker.WithLabels(map[string]string{
	// 	"RUNNING":       "example",
	// 	"IgnoreUnknown": "1",
	// }), worker.WithName(r.name+"-"+randSeq(4)), worker.WithSnapshot(r.snapshotName))
	// if err := node.Start(r.ctx); err != nil {
	// 	return err
	// }
	// r.workers <- node
	// r.downscaleMap[node.Name()] = time.AfterFunc(r.downscaleTimeout, func() {
	// 	r.killWorker(node)
	// })

	return nil
}

func (r *WorkerSet) Execute(ctx context.Context, payload []byte) (chan *ExecResult, error) {
	// res := make(chan *ExecResult)
	// w, err := r.retrieveWorker()
	// if err != nil {
	// 	return nil, err
	// }
	// go func() {
	// 	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(r.execTimeout)*time.Second)
	// 	defer cancel()

	// 	t := time.Now()
	// 	b, err := w.Execute(timeoutCtx, payload)
	// 	if err != nil {
	// 		log.Printf("[%s] timeout period elapsed %s", w.Name(), time.Since(t))
	// 		res <- &ExecResult{Err: err}
	// 		go r.killWorker(w)
	// 		return
	// 	}

	// 	log.Printf("[%s][%s] %s", w.Name(), time.Since(t), b)
	// 	defer func() {
	// 		r.downscaleMap[w.Name()].Reset(r.downscaleTimeout)
	// 		r.workers <- w
	// 	}()
	// 	res <- &ExecResult{Result: b}
	// }()

	// return res, nil
	return nil, nil
}

func (r *WorkerSet) retrieveWorker() (Worker, error) {
	// r.lock.Lock()
	// defer r.lock.Unlock()
	// if len(r.downscaleMap) < r.WorkerSetOptions.concurrency && len(r.workers) == 0 {
	// 	if err := r.Start(); err != nil {
	// 		return nil, err
	// 	}

	// }
	// w := <-r.workers
	// if r.downscaleMap[w.Name()] == nil {
	// 	return r.retrieveWorker()
	// }

	return nil, nil
}

func (r *WorkerSet) Shutdown() {
	// close(r.workers)

	// for w := range r.workers {
	// 	if r.downscaleMap[w.Name()] != nil {
	// 		w.Shutdown(r.ctx)
	// 	}
	// }
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
