package orchestration

import (
	"context"
	"sync"
)

type WorkersSet interface {
	Executor
	Shutdown(context.Context) error
}

type Executor interface {
	Execute(ctx context.Context, body []byte) (*ExecResult, error)
}

type WorkerSetManager struct {
	w sync.Map
}

func NewManager() *WorkerSetManager {
	return &WorkerSetManager{
		w: sync.Map{},
	}
}

func (r *WorkerSetManager) Register(ctx context.Context, name string, ws WorkersSet) {
	r.w.Store(name, ws)
}

func (r *WorkerSetManager) RetrieveWorker(name string) Executor {
	e, ok := r.w.Load(name)
	if !ok {
		return nil
	}
	return e.(Executor)
}

func (r *WorkerSetManager) Deregister(name string) {
	w, ok := r.w.LoadAndDelete(name)
	if ok {
		w.(WorkersSet).Shutdown(context.TODO())
	}
}

func (r *WorkerSetManager) Close() {
	r.w.Range(func(key, value any) bool {
		value.(WorkersSet).Shutdown(context.TODO())
		return true
	})
	r.w.Clear()
}
