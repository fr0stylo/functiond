package worker

import (
	"context"
	"log"

	containerd "github.com/containerd/containerd/v2/client"
)

type Worker interface {
	Start(context.Context) error
	Name() string
	Execute(context.Context, []byte) ([]byte, error)
	Shutdown(context.Context) error
}

type Node struct {
	*WorkerOptions
	runtime        *Runtime
	socket         *SocketHandler
	networkManager *NetworkManager
}

func NewNode(client *containerd.Client, nm *NetworkManager, opts ...WorkerOpts) (*Node, error) {
	options := defaultOptions
	for _, optFn := range opts {
		optFn(&options)
	}
	sh, err := NewSocketHandler(options.name)
	if err != nil {
		return nil, err
	}

	return &Node{
		WorkerOptions:  &options,
		runtime:        NewRuntime(client, &options),
		socket:         sh,
		networkManager: nm,
	}, nil
}

func (r *Node) Name() string {
	return r.name
}

func (r *Node) Execute(ctx context.Context, payload []byte) ([]byte, error) {
	return r.socket.Execute(payload), nil
}

func (r *Node) Shutdown(ctx context.Context) error {
	pid, err := r.runtime.GetPid()
	if err != nil {
		return err
	}
	if err := r.networkManager.Detach(ctx, pid, r.labels); err != nil {
		log.Print(err)
	}
	r.socket.Close()
	return r.runtime.Close(ctx)
}

func (r *Node) Start(ctx context.Context) error {
	socket := r.socket.Start()

	if err := r.runtime.Start(ctx, socket); err != nil {
		return err
	}

	pid, err := r.runtime.GetPid()
	if err != nil {
		return err
	}

	return r.networkManager.Attach(ctx, pid, r.labels)
}
