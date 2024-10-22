package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	Err_Timeout = errors.New("timeout")
)

type SocketHandler struct {
	socket    *net.UnixListener
	path      string
	message   chan []byte
	onMessage chan []byte
	cancel    chan struct{}
}

func NewSocketHandler(name string) (*SocketHandler, error) {
	socketPath := fmt.Sprintf("/etc/functiond/%s.sock", name)

	l, err := net.ListenUnix("unix", &net.UnixAddr{socketPath, "unix"})
	if err != nil {
		return nil, err
	}

	return &SocketHandler{
		socket:    l,
		path:      socketPath,
		message:   make(chan []byte),
		onMessage: make(chan []byte),
	}, nil
}

func (r *SocketHandler) Start() string {
	go func() {
		for {
			r.handleConnection()
		}
	}()

	return r.path
}

func (r *SocketHandler) handleConnection() {
	r.cancel = make(chan struct{})

	con, err := r.socket.AcceptUnix()
	if err != nil {
		log.Print(err)
		return
	}
	defer con.Close()
	body := <-r.message
	con.Write(body)
	buff := make([]byte, 0)
	b := bytes.NewBuffer(buff)
	if _, err := io.Copy(b, con); err != nil {
		log.Print(err)
		return
	}
	r.onMessage <- b.Bytes()
}

func (r *SocketHandler) Execute(ctx context.Context, payload []byte) ([]byte, error) {
	r.message <- payload

	select {
	case msg := <-r.onMessage:
		return msg, nil
	case <-ctx.Done():
		return nil, Err_Timeout
	}
}

func (r *SocketHandler) Close() {
	r.socket.Close()
}
