package tcp_mockserver

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestMockServer_New(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		m, err := New("127.0.0.1:0") // port 0 means auto
		if err != nil {
			t.Fatalf("New() returned unexpected error: %v", err)
		}
		defer m.Listener.Close()

		assert.Containsf(t, m.Listener.Addr().String(), "127.0.0.1:", "Bound address contains 127.0.0.1:")
		assert.Equalf(t, m.Listener.Addr().String(), m.Addr.String(), "Addr is set correctly")
		t.Logf("New() listener address: %s", m.Listener.Addr().String())
	})

	t.Run("error path: malformed address", func(t *testing.T) {
		_, err := New("127.0.0.1:abc")
		t.Logf("New() returned err: %v", err)
		assert.Errorf(t, err, "should return error")
	})

	t.Run("error path: failed to bind", func(t *testing.T) {
		_, err := New("8.8.8.8:1234")
		t.Logf("New() returned err: %v", err)
		assert.Errorf(t, err, "should return error")
	})
}

func Test_HandleConnection(t *testing.T) {
	t.Run("happy path: no sequences, with preamble", func(t *testing.T) {
		m, err := New("127.0.0.1:0") // port 0 means auto
		if err != nil {
			t.Fatalf("New() returned unexpected error: %v", err)
		}
		defer m.Listener.Close()
		m.Preamble = &ResponseSequence{ResponseData: []byte("foobar\n")}

		go func() {
			err = m.HandleConnection(1 * time.Second)
			if err != nil {
				t.Errorf("HandleConnection() returned error: %v", err)
			}
		}()

		conn, err := net.DialTCP("tcp", nil, m.Addr)
		if err != nil {
			t.Fatalf("connecting to mock failed with error: %v", err)
		}
		defer conn.Close()

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		assert.Equalf(t, nil, err, "should return no error")
		assert.Equalf(t, 7, n, "should return 7 bytes")
		assert.Equalf(t, []byte("foobar\n"), buffer[:n], "should return the right text")
	})

	t.Run("happy path: two sequences", func(t *testing.T) {

	})
}
