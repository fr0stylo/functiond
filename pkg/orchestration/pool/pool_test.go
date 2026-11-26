package pool_test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"functiond/pkg/orchestration/pool"
	"functiond/pkg/orchestration/pool/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type Element struct {
	data      string
	stopCalls int
}

func (e *Element) Stop(ctx context.Context) error {
	type ReleaseFunc func() error
	e.stopCalls++
	return nil
}

func TestPool(t *testing.T) {
	first := &Element{data: "one"}
	factory := mocks.NewMockPoolFactory[*Element](t)
	factory.EXPECT().Create(t.Context(), mock.Anything).Return(first, nil).Once()
	p := pool.NewPool(3, factory, 5*time.Second)

	created, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, created)
	createdValue, err := created.Value()
	require.NoError(t, err)
	assert.Same(t, first, createdValue)
	assert.Equal(t, 1, p.Size())
	assert.Zero(t, p.IdleSize())
	require.NoError(t, created.Release())

	require.NoError(t, p.Add(&Element{data: "two"}))
	require.NoError(t, p.Add(&Element{data: "three"}))

	l1, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, l1)
	l2, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, l2)
	l3, err := p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, l3)
	assert.Zero(t, p.IdleSize())
	assert.Equal(t, 3, p.Size())
	for _, check := range []struct {
		lease *pool.Lease[*Element]
		want  string
	}{
		{l1, "one"},
		{l2, "two"},
		{l3, "three"},
	} {
		value, err := check.lease.Value()
		require.NoError(t, err)
		require.NotNil(t, value)
		assert.Equal(t, check.want, value.data)
	}

	require.NoError(t, l2.Release())
	assert.Equal(t, 1, p.IdleSize())
	l2, err = p.Aquire(t.Context())
	require.NoError(t, err)
	require.NotNil(t, l2)
	assert.Zero(t, p.IdleSize())
	reused, err := l2.Value()
	require.NoError(t, err)
	require.NotNil(t, reused)
	assert.Equal(t, "two", reused.data)

	require.NoError(t, l3.Release())
	require.NoError(t, l1.Release())
	assert.Equal(t, 2, p.IdleSize())
	require.NoError(t, l2.Release())
	assert.Equal(t, 3, p.IdleSize())

	assert.Error(t, p.Add(&Element{data: "four"}))
}

func TestPoolReleaseMakesResourceAvailableAgain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := pool.NewPool[*Element](1, nil, 5*time.Second)
		want := &Element{data: "reusable"}
		require.NoError(t, p.Add(want))
		first, err := p.Aquire(t.Context())
		require.NoError(t, err)
		require.NotNil(t, first)
		firstValue, err := first.Value()
		require.NoError(t, err)
		assert.Same(t, want, firstValue)
		firstKey, err := first.Key()
		require.NoError(t, err)

		var second *pool.Lease[*Element]
		var acquireErr error
		acquired := false
		go func() {
			second, acquireErr = p.Aquire(t.Context())
			acquired = true
		}()
		synctest.Wait()
		assert.False(t, acquired, "acquire should wait while the only resource is leased")

		require.NoError(t, first.Release())
		synctest.Wait()
		require.True(t, acquired, "release should wake the waiting acquire")
		require.NoError(t, acquireErr)
		require.NotNil(t, second)
		secondValue, err := second.Value()
		require.NoError(t, err)
		assert.Same(t, want, secondValue)
		secondKey, err := second.Key()
		require.NoError(t, err)
		assert.Equal(t, firstKey, secondKey)
		require.NoError(t, second.Release())
		assert.Equal(t, 1, p.Size())
		assert.Equal(t, 1, p.IdleSize())
	})
}

func TestPoolAddWakesWaitingAcquire(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := pool.NewPool[*Element](1, nil, 5*time.Second)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		var lease *pool.Lease[*Element]
		var acquireErr error
		go func() {
			lease, acquireErr = p.Aquire(ctx)
		}()
		synctest.Wait()
		require.Nil(t, lease)

		want := &Element{data: "added"}
		require.NoError(t, p.Add(want))
		synctest.Wait()
		require.NoError(t, acquireErr)
		require.NotNil(t, lease)
		value, err := lease.Value()
		require.NoError(t, err)
		assert.Same(t, want, value)
		require.NoError(t, lease.Release())
	})
}

func TestPoolWaitingAcquireReturnsOnCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := pool.NewPool[*Element](1, nil, 5*time.Second)
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		var lease *pool.Lease[*Element]
		var acquireErr error
		go func() {
			lease, acquireErr = p.Aquire(ctx)
		}()
		synctest.Wait()
		require.NoError(t, acquireErr)

		cancel()
		synctest.Wait()
		require.ErrorIs(t, acquireErr, context.Canceled)
		require.Nil(t, lease)
	})
}

func TestPoolDestroyRemovesResourceAndStopsIt(t *testing.T) {
	p := pool.NewPool[*Element](1, nil, 5*time.Second)
	resource := &Element{data: "temporary"}
	assert.NoError(t, p.Add(resource))

	lease, err := p.Aquire(t.Context())
	assert.NoError(t, err)
	if !assert.NotNil(t, lease) {
		return
	}
	require.NoError(t, lease.Destroy(context.Background()))

	assert.Equal(t, 0, p.Size())
	assert.Equal(t, 0, p.IdleSize())
	assert.Equal(t, 1, resource.stopCalls)

	// Destroying a resource frees its capacity for a replacement.
	assert.NoError(t, p.Add(&Element{data: "replacement"}))
	assert.Equal(t, 1, p.Size())
}

func TestPoolDestroyLetsWaitingAcquireCreateReplacement(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		resource := &Element{data: "temporary"}
		replacement := &Element{data: "replacement"}
		factory := mocks.NewMockPoolFactory[*Element](t)
		factory.EXPECT().Create(t.Context(), mock.Anything).Return(replacement, nil).Once()
		p := pool.NewPool(1, factory, 5*time.Second)
		require.NoError(t, p.Add(resource))
		first, err := p.Aquire(t.Context())
		require.NoError(t, err)
		require.NotNil(t, first)

		var second *pool.Lease[*Element]
		var acquireErr error
		go func() {
			second, acquireErr = p.Aquire(t.Context())
		}()
		synctest.Wait()
		require.NoError(t, first.Destroy(t.Context()))
		synctest.Wait()

		require.NoError(t, acquireErr)
		require.NotNil(t, second, "destroy should let the waiting acquire create a replacement")
		value, err := second.Value()
		require.NoError(t, err)
		assert.Same(t, replacement, value)
		assert.Equal(t, 1, resource.stopCalls)
		assert.Equal(t, 1, p.Size())
		require.NoError(t, second.Release())
	})
}
