package pool

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"
)

type Stopable interface {
	Stop(ctx context.Context) error
}

type PoolFactory[V Stopable] interface {
	Create(ctx context.Context, name string) (V, error)
}

type Pool[V Stopable] struct {
	m                *sync.Mutex
	s                *sync.Cond
	idleQueue        *list.List
	container        map[string]*Element[V]
	maxSize          int
	factory          PoolFactory[V]
	cleanerInterval  time.Duration
	downscaleTimeout time.Duration
}

func NewPool[V Stopable](maxSize int, poolFactory PoolFactory[V], downscaleTimeout time.Duration) *Pool[V] {
	m := &sync.Mutex{}
	p := &Pool[V]{
		idleQueue:        list.New(),
		container:        map[string]*Element[V]{},
		m:                m,
		s:                sync.NewCond(m),
		maxSize:          maxSize,
		factory:          poolFactory,
		cleanerInterval:  5 * time.Second,
		downscaleTimeout: 5 * time.Second,
	}

	time.AfterFunc(p.cleanerInterval, p.downscale)
	return p
}

func (r *Pool[V]) formLease(el *Element[V]) *Lease[V] {
	r.container[el.name] = el
	el.refCount++

	return &Lease[V]{
		resource: el,
		key:      el.name,
		pool:     r,
	}
}

func (r *Pool[V]) downscale() {
	r.m.Lock()
	defer r.m.Unlock()
	t := time.Now().UnixMilli()
	leases := []*Lease[V]{}

	for e := r.idleQueue.Back(); e != nil && t-e.Value.(*Element[V]).touched > r.downscaleTimeout.Milliseconds(); e = r.idleQueue.Back() {
		el := e.Value.(*Element[V])
		leases = append(leases, r.formLease(el))
		r.idleQueue.Remove(e)
	}

	for _, l := range leases {
		go l.Destroy(context.Background())
	}

	time.AfterFunc(r.cleanerInterval, r.downscale)
}

// Aquire returns an idle resource or creates one when there is room and a factory.
// Otherwise it waits for a resource to be released or added, or for room to create one.
func (r *Pool[V]) Aquire() (*Lease[V], error) {
	r.m.Lock()
	defer r.m.Unlock()

	for {
		if e := r.idleQueue.Front(); e != nil {
			r.idleQueue.Remove(e)
			return r.formLease(e.Value.(*Element[V])), nil
		}

		if r.factory != nil && len(r.container) < r.maxSize {
			e, err := NewElementFromFactory(r.factory)
			if err != nil {
				// Capacity is still available for another waiter.
				r.s.Signal()
				return nil, err
			}
			return r.formLease(e), nil
		}

		r.s.Wait()
	}
}

func (r *Pool[V]) Add(el V) error {
	r.m.Lock()
	defer r.m.Unlock()

	if len(r.container) >= r.maxSize {
		return errors.New("Max capacity reached")
	}

	e := NewElement(el)
	r.idleQueue.PushBack(e)
	r.container[e.name] = e
	r.s.Signal()

	return nil
}

func (r *Pool[V]) IdleSize() int {
	r.m.Lock()
	defer r.m.Unlock()

	return r.idleQueue.Len()
}

func (r *Pool[V]) Size() int {
	r.m.Lock()
	defer r.m.Unlock()

	return len(r.container)
}

func (r *Pool[V]) Close(ctx context.Context) error {
	for _, v := range r.container {
		v.val.Stop(ctx)
	}

	return nil
}
