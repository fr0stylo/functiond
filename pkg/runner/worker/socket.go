package worker

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type SocketHandler struct {
	socket    *net.UnixListener
	path      string
	message   chan []byte
	onMessage chan []byte
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
			con, err := r.socket.AcceptUnix()
			if err != nil {
				log.Fatal(err)
			}
			body := <-r.message

			con.Write(body)
			buff := make([]byte, 0)
			b := bytes.NewBuffer(buff)
			if _, err := io.Copy(b, con); err != nil {
				log.Print(err)
			}

			r.onMessage <- b.Bytes()
			con.Close()
		}
	}()

	return r.path
}

func (r *SocketHandler) Execute(payload []byte) []byte {
	r.message <- payload
	return <-r.onMessage
}

func (r *SocketHandler) Close() {
	r.socket.Close()
	os.ReadFile(r.path)
}
