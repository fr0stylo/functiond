package pool_test

import (
	"context"
	"testing"
	"time"

	"functiond/pkg/orchestration/pool"

	"github.com/stretchr/testify/assert"
)

type leaseResource struct {
	stopCalls int
}

func (r *leaseResource) Stop(context.Context) error {
	r.stopCalls++
	return nil
}

func TestLeaseValueAndKey(t *testing.T) {
	p := pool.NewPool[*leaseResource](1, nil, 5*time.Second)
	want := &leaseResource{}
	assert.NoError(t, p.Add(want))

	lease, err := p.Aquire()
	assert.NoError(t, err)
	if assert.NotNil(t, lease) {
		assert.Same(t, want, lease.Value())
		assert.NotEmpty(t, lease.Key())
	}
}

func TestLeaseReleaseReturnsResourceToPool(t *testing.T) {
	p := pool.NewPool[*leaseResource](1, nil, 5*time.Second)
	want := &leaseResource{}
	assert.NoError(t, p.Add(want))

	lease, err := p.Aquire()
	assert.NoError(t, err)
	if !assert.NotNil(t, lease) {
		return
	}
	lease.Release()

	assert.Equal(t, 1, p.IdleSize())
	reused, err := p.Aquire()
	assert.NoError(t, err)
	if assert.NotNil(t, reused) {
		assert.Same(t, want, reused.Value())
	}
}

func TestLeaseDestroyStopsResource(t *testing.T) {
	p := pool.NewPool[*leaseResource](1, nil, 5*time.Second)
	resource := &leaseResource{}
	assert.NoError(t, p.Add(resource))

	lease, err := p.Aquire()
	assert.NoError(t, err)
	if !assert.NotNil(t, lease) {
		return
	}
	lease.Destroy(context.Background())

	assert.Equal(t, 1, resource.stopCalls)
	assert.Equal(t, 0, p.Size())
}
