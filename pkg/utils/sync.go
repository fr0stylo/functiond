package utils

type Notifier struct {
	n chan struct{}
}

func NewNotifier() *Notifier {
	return &Notifier{
		n: make(chan struct{}, 1),
	}
}

func (r *Notifier) Wait() <-chan struct{} {
	return r.n
}

func (r *Notifier) Signal() {
	select {
	case r.n <- struct{}{}:
	default:
	}
}

func (r *Notifier) Notify() {
	close(r.n)
	r.n = make(chan struct{}, 1)
}
