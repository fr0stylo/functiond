package worker

import (
	"context"
	"fmt"
	"log"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Runtime struct {
	*WorkerOptions
	client    *containerd.Client
	task      containerd.Task
	container containerd.Container
}

func NewRuntime(client *containerd.Client, wo *WorkerOptions) *Runtime {
	return &Runtime{client: client, WorkerOptions: wo}
}

func (r *Runtime) GetPid() (uint32, error) {
	if r.task == nil {
		return 0, fmt.Errorf("task not started")
	}
	return r.task.Pid(), nil
}

func (r *Runtime) Start(ctx context.Context, socketPath string) error {
	image, err := r.client.GetImage(ctx, "docker.io/library/node:lts-alpine")
	if err != nil {
		return fmt.Errorf("failed to get image: %v", err)
	}

	container, err := r.client.LoadContainer(ctx, r.name)
	if err != nil {
		snapSvc := r.client.SnapshotService(defaults.DefaultSnapshotter)

		_, err := snapSvc.Prepare(ctx, r.name, r.snapshotName)
		if err != nil {
			return err
		}

		container, err = r.client.NewContainer(
			ctx,
			r.name,
			containerd.WithSnapshot(r.name),
			containerd.WithNewSpec(
				oci.WithImageConfig(image),
				oci.WithRootFSReadonly(),
				oci.WithHostResolvconf,
				oci.WithNoNewPrivileges,
				oci.WithHostname(r.name),
				oci.WithCPUShares(128),
				oci.WithMemoryLimit(128*1024*1024),
				oci.WithProcessCwd("/opt/function/lambda"),
				oci.WithProcessArgs("node", "lambda.js"),
				oci.WithMounts([]specs.Mount{
					{
						Type:        "bind",
						Destination: "/etc/functiond/functiond.sock",
						Source:      socketPath,
						Options:     []string{"rbind", "rw"},
					},
				}),
			),
		)
		if err != nil {
			return err
		}
	}
	r.container = container

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return err
	}
	r.task = task

	if err := task.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Runtime) Close(ctx context.Context) error {
	log.Print("Waiting to kill")
	status, err := r.task.Delete(ctx, containerd.WithProcessKill)

	log.Printf("Killed %d", status.ExitCode())

	r.container.Delete(ctx, containerd.WithSnapshotCleanup)
	return err
}
