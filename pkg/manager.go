package pkg

import (
	"context"

	"functiond/pkg/runner"
)

type WorkerSetManager struct {
	workerSets map[string]*runner.WorkerSet
}

func NewManager() *WorkerSetManager {
	return &WorkerSetManager{
		workerSets: map[string]*runner.WorkerSet{},
	}
}

func (r *WorkerSetManager) Register(ctx context.Context, opts *runner.WorkerSetOptions) error {
	ws, err := runner.NewWorkerSet(ctx, runner.WithOptions(opts))
	if err != nil {
		return err
	}
	r.workerSets[ws.Name()] = ws

	return nil
}

func (r *WorkerSetManager) RegisterVersion() {

}

func (r *WorkerSetManager) RetrieveWorker(name string) *runner.WorkerSet {
	return r.workerSets[name]
}

func (r *WorkerSetManager) Deregister(name string) {
	r.workerSets[name].Shutdown()
	delete(r.workerSets, name)
}

func (r *WorkerSetManager) Close() {
	for k := range r.workerSets {
		r.Deregister(k)
	}
}
