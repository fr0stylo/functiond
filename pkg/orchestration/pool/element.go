package pool

import (
	"context"
	"functiond/pkg/utils"
	"time"
)

type Element[V Stopable] struct {
	name     string
	val      V
	refCount int
	touched  int64
}

func NewElement[V Stopable](v V) *Element[V] {
	return &Element[V]{
		name:     utils.RandStringN(16),
		val:      v,
		refCount: 0,
		touched:  time.Now().UnixMilli(),
	}
}

func NewElementFromFactory[V Stopable](ctx context.Context, v PoolFactory[V]) (*Element[V], error) {
	name := utils.RandStringN(16)
	e, err := v.Create(ctx, name)
	if err != nil {
		return nil, err
	}
	return &Element[V]{
		name:     name,
		val:      e,
		refCount: 0,
		touched:  time.Now().UnixMilli(),
	}, nil
}

func (r *Element[V]) Touch() {
	r.touched = time.Now().UnixMilli()
}
