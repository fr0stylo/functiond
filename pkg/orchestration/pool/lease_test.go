package pool_test

import (
	"context"
	"testing"
	"time"

	"functiond/pkg/orchestration/pool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, p.Add(want))

	lease, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, lease)

	value, err := lease.Value()
	require.NoError(t, err)
	assert.Same(t, want, value)
	key, err := lease.Key()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
}

func TestLeaseReleaseReturnsResourceToPool(t *testing.T) {
	p := pool.NewPool[*leaseResource](1, nil, 5*time.Second)
	want := &leaseResource{}
	require.NoError(t, p.Add(want))

	lease, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, lease)
	require.NoError(t, lease.Release())
	assert.ErrorIs(t, lease.Release(), pool.ErrorLeaseExpired)

	value, err := lease.Value()
	assert.ErrorIs(t, err, pool.ErrorLeaseExpired)
	assert.Nil(t, value)
	key, err := lease.Key()
	assert.ErrorIs(t, err, pool.ErrorLeaseExpired)
	assert.Empty(t, key)

	assert.Equal(t, 1, p.IdleSize())
	reused, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, reused)
	value, err = reused.Value()
	require.NoError(t, err)
	assert.Same(t, want, value)
}

func TestLeaseDestroyStopsResource(t *testing.T) {
	p := pool.NewPool[*leaseResource](1, nil, 5*time.Second)
	resource := &leaseResource{}
	require.NoError(t, p.Add(resource))

	lease, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, lease)
	require.NoError(t, lease.Destroy(t.Context()))
	assert.ErrorIs(t, lease.Release(), pool.ErrorLeaseExpired)
	assert.ErrorIs(t, lease.Destroy(t.Context()), pool.ErrorLeaseExpired)

	assert.Equal(t, 1, resource.stopCalls)
	assert.Equal(t, 0, p.Size())
}
