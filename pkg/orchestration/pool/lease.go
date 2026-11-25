package pool

import (
	"context"
)

type Lease[V Stopable] struct {
	pool     *Pool[V]
	key      string
	resource *Element[V]
}

func (r *Lease[V]) Release() {
	r.pool.m.Lock()
	defer r.pool.m.Unlock()

	r.resource.refCount -= 1
	r.pool.idleQueue.PushBack(r.resource)
	r.pool.s.Signal()
}

func (r *Lease[V]) Destroy(ctx context.Context) {
	r.pool.m.Lock()
	delete(r.pool.container, r.resource.name)
	r.pool.m.Unlock()

	el := r.resource.val
	el.Stop(ctx)
	r.pool.s.Signal()
}

func (r *Lease[V]) Value() V {
	return r.resource.val
}

func (r *Lease[V]) Key() string {
	return r.resource.name
}
