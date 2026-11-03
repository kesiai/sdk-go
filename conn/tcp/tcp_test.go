package tcp

import (
	"io"
	"net"
	"testing"
	"time"
)

func Test_Read(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { listener.Close() })
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			done <- err
			return
		}
		_, err = io.Copy(conn, conn)
		done <- err
	}()
	conn, err := DialTCP("tcp", "127.0.0.1", listener.Addr().(*net.TCPAddr).Port)
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			conn.Close()
		}
	})
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	want := "sdk-go echo"
	if _, err := conn.Write([]byte(want)); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(want))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != want {
		t.Fatalf("echo = %q, want %q", buf, want)
	}
	conn.Close()
	closed = true
	if _, err := conn.Write([]byte(want)); err == nil {
		t.Fatal("Write after Close succeeded")
	}
	if _, err := conn.Read(buf); err == nil {
		t.Fatal("Read after Close succeeded")
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
