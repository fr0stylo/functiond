package shell

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"functiond/pkg/orchestration"
)

// Shell runs each payload as a separate shell command.
type Shell struct {
	name  string
	shell string
	dir   string
	env   []string

	mu      sync.Mutex
	started bool
}

var _ orchestration.Worker = (*Shell)(nil)

func (worker *Shell) Name() string { return worker.name }

func (worker *Shell) Init(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if worker.shell == "" {
		return fmt.Errorf("shell executable must not be empty")
	}
	if _, err := exec.LookPath(worker.shell); err != nil {
		return fmt.Errorf("find shell executable %q: %w", worker.shell, err)
	}
	return nil
}

func (worker *Shell) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	worker.mu.Lock()
	defer worker.mu.Unlock()
	if worker.started {
		return fmt.Errorf("shell worker %q is already started", worker.name)
	}
	worker.started = true
	return nil
}

func (worker *Shell) Execute(ctx context.Context, payload []byte) (*orchestration.ExecResult, error) {
	worker.mu.Lock()
	started := worker.started
	worker.mu.Unlock()
	if !started {
		return nil, fmt.Errorf("shell worker %q is not started", worker.name)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(payload))) == 0 {
		return nil, fmt.Errorf("shell payload must not be empty")
	}

	cmd := exec.CommandContext(ctx, worker.shell, "-c", string(payload))
	cmd.Dir = worker.dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SYS_KILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	if worker.env != nil {
		cmd.Env = worker.env
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return &orchestration.ExecResult{
			Result: output,
			Err:    fmt.Errorf("run shell payload: %w", err),
		}, nil
	}
	return &orchestration.ExecResult{Result: output}, nil
}

func (worker *Shell) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	worker.mu.Lock()
	worker.started = false
	worker.mu.Unlock()
	fmt.Printf("Killing worker %s\n", worker.name)
	return nil
}
