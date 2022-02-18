package tcp_mockserver

import (
	"github.com/pkg/errors"
	"net"
	"time"
)

type MockServer struct {
	Listener          *net.TCPListener  // The socket used for listening
	Addr              *net.TCPAddr      // the address we're listening on, for use by the client
	Preamble          *ResponseSequence // what to send after the connection is established
	ResponseSequences []*ResponseSequence
}

type ResponseSequence struct {
	ListenTimeout  time.Duration         // how long to wait for data
	Delay          time.Duration         // how long to wait before sending this response
	ResponseData   []byte                // the data to send as response
	Test           func(received []byte) // function to test the data received by the mock
	ReadBufferSize int                   // how large the read buffer should be (defaults to 1024
}

func New(addr string) (*MockServer, error) {
	listenAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, errors.Wrapf(err, "tcp-mockserver.New: parsing address '%s'", addr)
	}

	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return nil, errors.Wrap(err, "tcp-mockserver.New: setting up listener")
	}

	tcpAddr, _ := net.ResolveTCPAddr("tcp", listener.Addr().String())
	return &MockServer{Listener: listener, Addr: tcpAddr}, nil
}

// HandleConnection will wait for a connection, send the preamble and start the conversation
func (M *MockServer) HandleConnection(timeout time.Duration) error {
	err := M.Listener.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return errors.Wrap(err, "tcp-mockserver.HandleConnection: waiting for new connection")
	}

	conn, err := M.Listener.Accept()
	if err != nil {
		return errors.Wrap(err, "tcp-mockserver.HandleConnection: accepting connection")
	}
	defer conn.Close()

	if M.Preamble != nil {
		time.Sleep(M.Preamble.Delay)
		_, err = conn.Write(M.Preamble.ResponseData)
		if err != nil {
			return errors.Wrap(err, "tcp-mockserver.HandleConnection: sending preamble")
		}

	}

	for _, responseSequence := range M.ResponseSequences {
		err := M.Respond(conn, responseSequence)
		if err != nil {
			return errors.Wrapf(err, "tcp-mockserver.HandleConnection: responding with sequence: %+v", responseSequence)
		}
	}

	return nil
}

// Respond will wait for incoming data and respond to it
func (M *MockServer) Respond(conn net.Conn, response *ResponseSequence) error {
	err := conn.SetReadDeadline(time.Now().Add(response.ListenTimeout))
	if err != nil {
		return errors.Wrap(err, "tcp-mockserver.Respond: setting deadline for read")
	}

	bufferSize := response.ReadBufferSize
	if bufferSize == 0 {
		bufferSize = 1024
	}

	data := make([]byte, bufferSize)
	n, err := conn.Read(data)
	if err != nil {
		return errors.Wrap(err, "tcp-mockserver.Respond: reading data")
	}

	if response.Test != nil {
		response.Test(data[:n])
	}

	if response.ResponseData == nil {
		return nil // we're done
	}

	time.Sleep(response.Delay)
	_, err = conn.Write(response.ResponseData)
	if err != nil {
		return errors.Wrap(err, "tcp-mockserver.Respond: sending response")
	}

	return nil
}
