package pool

import (
	"context"
	"errors"
)

var (
	ErrorLeaseExpired error = errors.New("lease was expired, was released or destroyed")
)

type Lease[V Stopable] struct {
	pool     *Pool[V]
	key      string
	resource *Element[V]
}

func (r *Lease[V]) Release() error {
	if r.resource == nil {
		return ErrorLeaseExpired
	}

	r.pool.m.Lock()
	defer r.pool.m.Unlock()

	r.resource.refCount -= 1
	r.pool.idleQueue.PushBack(r.resource)
	r.pool.s.Signal()
	r.resource = nil

	return nil
}

func (r *Lease[V]) Destroy(ctx context.Context) error {
	if r.resource == nil {
		return ErrorLeaseExpired
	}
	r.pool.m.Lock()
	delete(r.pool.container, r.resource.name)
	r.pool.m.Unlock()

	el := r.resource.val
	el.Stop(ctx)
	r.pool.s.Signal()
	r.resource = nil

	return nil
}

func (r *Lease[V]) Value() (V, error) {
	if r.resource == nil {
		var v V
		return v, ErrorLeaseExpired
	}
	return r.resource.val, nil
}

func (r *Lease[V]) Key() (string, error) {
	if r.resource == nil {
		return "", ErrorLeaseExpired
	}

	return r.resource.name, nil
}
