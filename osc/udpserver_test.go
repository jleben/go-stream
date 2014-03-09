package osc

import (
	"net"
	"testing"
	"time"
)

func newTestServer(timeout time.Duration) *UDPServer {
	return &UDPServer{
		Address:     ":11111",
		ReadTimeout: timeout,
	}
}

func TestListenAndServeTimeout(t *testing.T) {

	srv := newTestServer(100 * time.Millisecond)

	// ensure we timeout as expected
	if err := srv.ListenAndServe(); !err.(net.Error).Timeout() {
		t.Error(err)
	}
}

type OSCTestReceiver struct {
	cm chan *Message
}

func (r *OSCTestReceiver) HandleOSC(m *Message) {
	r.cm <- m
}

func TestListenAndServe(t *testing.T) {

	cm := make(chan *Message, 1)

	messageHandler1 := func(m *Message) {
		cm <- m
	}

	recvr := &OSCTestReceiver{cm}

	// register handlers in both ways
	HandleFunc("/test1", messageHandler1)
	Handle("/test2", recvr)

	srv := newTestServer(0)

	go srv.ListenAndServe()

	// Connect to server
	raddr, err := net.ResolveUDPAddr("udp", srv.Address)
	if err != nil {
		t.Error("ResolveUDPAddr:", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		t.Error("DialUDP:", err)
	}
	defer conn.Close()

	// Create OSC Message
	m1 := &Message{Address: "/test*"}
	m1.Args = append(m1.Args, float32(3.1415))

	// XXX: ensure the server has time to start listening
	time.Sleep(10 * time.Millisecond)

	// Write to server
	_, err = m1.WriteTo(conn)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	keepgoing := true
	var msgs []*Message

	for keepgoing {
		select {
		case msg := <-cm:
			msgs = append(msgs, msg)
			if len(msgs) >= 2 {
				keepgoing = false
			}
		case <-time.After(1 * time.Second):
			t.Errorf("Timed out. got %d msgs, wanted %d", len(msgs), 2)
			keepgoing = false
		}
	}

	for _, m := range msgs {
		if !Equal(m1, m) {
			t.Error("received message didn't match:", m1, m)
		}
	}
}
