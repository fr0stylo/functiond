package worker

import gocni "github.com/containerd/go-cni"

type WorkerOpts = func(opts *WorkerOptions)
type Opts[T any] func(opts *T)
type WorkerOptions struct {
	labels       map[string]string
	portMapping  []gocni.PortMapping
	snapshotName string
	name         string
}

var defaultOptions = WorkerOptions{
	labels:      map[string]string{},
	portMapping: make([]gocni.PortMapping, 0),
}

type PortMapping = gocni.PortMapping

func WithPortMapping(pm []PortMapping) WorkerOpts {
	return func(opts *WorkerOptions) {
		opts.portMapping = pm
	}
}

func WithLabels(labels map[string]string) WorkerOpts {
	return func(opts *WorkerOptions) {
		opts.labels = labels
	}
}

func WithSnapshot(snapshot string) WorkerOpts {
	return func(opts *WorkerOptions) {
		opts.snapshotName = snapshot
	}
}

func WithName(name string) WorkerOpts {
	return func(opts *WorkerOptions) {
		opts.name = name
	}
}
