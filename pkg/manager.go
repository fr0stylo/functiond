package pkg

import (
	"context"

	"functiond/pkg/runner"
	"functiond/pkg/runner/worker"
)

type Manager struct {
	workerSets map[string]*runner.WorkerSet
}

func NewManager() *Manager {
	return &Manager{
		workerSets: map[string]*runner.WorkerSet{},
	}
}

func (r *Manager) Register(ctx context.Context, opts ...worker.Opts[runner.WorkerSetOptions]) error {
	ws, err := runner.NewWorkerSet(ctx, opts...)
	if err != nil {
		return err
	}
	r.workerSets[ws.Name()] = ws

	return nil
}

func (r *Manager) RegisterVersion() {

}

func (r *Manager) RetrieveWorker(name string) *runner.WorkerSet {
	return r.workerSets[name]
}

func (r *Manager) Deregister(name string) {
	r.workerSets[name].Shutdown()
	delete(r.workerSets, name)
}

func (r *Manager) Close() {
	for k := range r.workerSets {
		r.Deregister(k)
	}
}
