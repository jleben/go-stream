package osc

import (
	// "fmt"
	"net"
	"time"
)

type Dispatcher interface {
	Dispatch(*Bundle)
}

type Handler interface {
	HandleOSC(*Message)
}

type HandlerFunc func(*Message)

func (f HandlerFunc) HandleOSC(m *Message) {
	f(m)
}

type ServeMux struct {
	handlers map[string]Handler
}

func NewServeMux() *ServeMux {
	return &ServeMux{handlers: make(map[string]Handler)}
}

func (mux *ServeMux) Dispatch(b *Bundle) {

	// XXX: deliver according to timestamp

	for _, msg := range b.Messages {
		for address, handler := range mux.handlers {
			if PatternMatch(msg.Address, address) {
				handler.HandleOSC(msg)
			}
		}
	}
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	mux.handlers[pattern] = handler
}

var defaultServeMux = NewServeMux()

func Handle(pattern string, h Handler) {
	defaultServeMux.Handle(pattern, h)
}

func HandleFunc(pattern string, handler func(*Message)) {
	defaultServeMux.Handle(pattern, HandlerFunc(handler))
}

type UDPServer struct {
	Address     string        // UDP address to listen on
	Dispatcher  Dispatcher    // Dispatcher to invoke, osc.defaultServeMux if nil
	ReadTimeout time.Duration // maximum duration before timing out read of the request
}

func (s *UDPServer) ListenAndServe() error {

	if s.Dispatcher == nil {
		s.Dispatcher = defaultServeMux
	}

	udpAddr, err := net.ResolveUDPAddr("udp", s.Address)
	if err != nil {
		return err
	}

	for {
		conn, listenerr := net.ListenUDP("udp", udpAddr)
		if listenerr != nil {
			return listenerr
		}

		if s.ReadTimeout != 0 {
			conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
		}

		bundle, bndlerr := ReadFrom(conn)
		closerr := conn.Close()

		if bndlerr != nil {
			return bndlerr
		}
		if closerr != nil {
			return closerr
		}

		// XXX: dispatch in a goroutine?
		s.Dispatcher.Dispatch(bundle)
	}

	panic("unreachable")
}

func ListenAndServeUDP(addr string, dispatcher Dispatcher) error {
	server := &UDPServer{Address: addr, Dispatcher: dispatcher}
	return server.ListenAndServe()
}
