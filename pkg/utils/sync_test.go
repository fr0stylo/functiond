package utils_test

import (
	"testing"
	"testing/synctest"

	"functiond/pkg/utils"
)

func TestNotifierWaitReturnsCurrentChannel(t *testing.T) {
	notifier := utils.NewNotifier()
	wait := notifier.Wait()
	requireIdleNotifierChannel(t, wait)
	if got := notifier.Wait(); got != wait {
		t.Fatal("Wait returned a different channel before Notify")
	}
}

func TestNotifierInstancesAreIndependent(t *testing.T) {
	first := utils.NewNotifier()
	second := utils.NewNotifier()
	secondWait := second.Wait()
	if first.Wait() == secondWait {
		t.Fatal("separate notifiers share a wait channel")
	}

	first.Signal()
	requireIdleNotifierChannel(t, secondWait)
	first.Notify()
	requireIdleNotifierChannel(t, secondWait)
	if got := second.Wait(); got != secondWait {
		t.Fatal("notifying another instance changed the wait channel")
	}
}

func TestNotifierSignalBuffersAndCoalesces(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		notifier := utils.NewNotifier()
		wait := notifier.Wait()
		for range 10 {
			notifier.Signal()
		}
		if got := notifier.Wait(); got != wait {
			t.Fatal("Signal replaced the wait channel")
		}
		requireNotifierReceive(t, wait, true)
		requireIdleNotifierChannel(t, wait)

		notifier.Signal()
		requireNotifierReceive(t, wait, true)
		requireIdleNotifierChannel(t, wait)
	})
}

func TestNotifierConcurrentSignalsCoalesce(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const senders = 32
		notifier := utils.NewNotifier()
		wait := notifier.Wait()
		start := make(chan struct{})
		completed := make(chan (<-chan struct{}), senders)
		for range senders {
			go func() {
				<-start
				current := notifier.Wait()
				notifier.Signal()
				completed <- current
			}()
		}

		close(start)
		synctest.Wait()
		if got := len(completed); got != senders {
			t.Fatalf("%d of %d senders returned without a receiver", got, senders)
		}
		for range senders {
			if got := <-completed; got != wait {
				t.Fatal("concurrent Wait returned a different channel")
			}
		}
		requireNotifierReceive(t, wait, true)
		requireIdleNotifierChannel(t, wait)
	})
}

func TestNotifierNotifyReplacesChannel(t *testing.T) {
	for _, tc := range []struct {
		name          string
		pendingSignal bool
	}{
		{name: "idle"},
		{name: "pending signal", pendingSignal: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			notifier := utils.NewNotifier()
			first := notifier.Wait()
			if tc.pendingSignal {
				notifier.Signal()
			}

			notifier.Notify()
			second := notifier.Wait()
			if second == first {
				t.Fatal("Notify reused the previous wait channel")
			}
			if tc.pendingSignal {
				requireNotifierReceive(t, first, true)
			}
			requireNotifierReceive(t, first, false)
			requireIdleNotifierChannel(t, second)

			notifier.Notify()
			third := notifier.Wait()
			if third == first || third == second {
				t.Fatal("repeated Notify reused a previous wait channel")
			}
			requireNotifierReceive(t, first, false)
			requireNotifierReceive(t, second, false)
			requireIdleNotifierChannel(t, third)

			notifier.Signal()
			requireNotifierReceive(t, third, true)
			requireIdleNotifierChannel(t, third)
		})
	}
}

func TestNotifierSignalWakesOneWaiterAndNotifyWakesTheRest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const waiters = 4
		notifier := utils.NewNotifier()
		results := make(chan bool, waiters)
		stop := make(chan struct{})
		defer close(stop)
		for range waiters {
			go func() {
				select {
				case _, ok := <-notifier.Wait():
					results <- ok
				case <-stop:
				}
			}()
		}
		synctest.Wait()
		if got := len(results); got != 0 {
			t.Fatalf("%d waiters woke before any signal", got)
		}

		notifier.Signal()
		synctest.Wait()
		if got := len(results); got != 1 {
			t.Fatalf("Signal woke %d waiters, want 1", got)
		}
		if ok := <-results; !ok {
			t.Fatal("Signal closed the wait channel")
		}

		notifier.Notify()
		synctest.Wait()
		if got := len(results); got != waiters-1 {
			t.Fatalf("Notify woke %d remaining waiters, want %d", got, waiters-1)
		}
		for range waiters - 1 {
			if ok := <-results; ok {
				t.Fatal("waiter received a value after Notify, want channel closure")
			}
		}
		requireIdleNotifierChannel(t, notifier.Wait())
	})
}

func requireIdleNotifierChannel(t *testing.T, wait <-chan struct{}) {
	t.Helper()
	if wait == nil {
		t.Fatal("Wait returned a nil channel")
	}
	select {
	case _, ok := <-wait:
		t.Fatalf("wait channel was ready with receive ok = %t, want an open, empty channel", ok)
	default:
	}
}

func requireNotifierReceive(t *testing.T, wait <-chan struct{}, wantOK bool) {
	t.Helper()
	select {
	case _, ok := <-wait:
		if ok != wantOK {
			t.Fatalf("receive from wait channel returned ok = %t, want %t", ok, wantOK)
		}
	default:
		t.Fatalf("wait channel had no ready receive, want receive with ok = %t", wantOK)
	}
}
