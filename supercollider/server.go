package supercollider

import (
  "net"
  "../osc"
)

type sc_error struct {
  msg string
}

func (err sc_error) Error () string { return err.msg }

type Server struct {
  address *net.UDPAddr
  connection *net.UDPConn
  synth_ids int
}

func NewServer(address string) (*Server, error) {
  server := new(Server)
  server.synth_ids = 1000
  var err error;
  server.address, err = net.ResolveUDPAddr("udp", address)
  if err != nil { server = nil }
  return server, err
}

func (server *Server) Connect () error {
  if server.connection != nil {
    return sc_error { "Can not make new connection to server: already connected." }
  }
  var err error
  server.connection, err = net.DialUDP("udp", nil, server.address)
  return err
}

func (server *Server) Disconnect () error {
  if server.connection == nil {
    return sc_error { "Can not disconnect from server: not connected." }
  }
  var err error
  err = server.connection.Close()
  server.connection = nil
  return err
}

func (server *Server) SendMessage (msg *osc.Message) error {
  if server.connection == nil {
    return sc_error { "Can not send message: not connected." }
  }
  _, err := msg.WriteTo(server.connection)
  return err
}

func (server *Server) DumpOSC (on bool) error {
  msg := &osc.Message{Address: "/dumpOSC"}
  var arg int32
  if (on) { arg = 1 } else { arg = 0 }
  msg.Args = append(msg.Args, arg)
  err := server.SendMessage(msg)
  return err
}

func (server *Server) DumpTree (id int32) error {
  msg := &osc.Message{Address: "/g_dumpTree"}
  msg.Args = append(msg.Args, id)
  msg.Args = append(msg.Args, int32(0))
  e := server.SendMessage(msg)
  return e
}

func (server *Server) NewSynth (name string, params ... interface {}) (id int32, err error) {
  if len(params) % 2 != 0 {
    return -1, sc_error { "Can not create synth: odd number of parameters." }
  }

  msg := &osc.Message{Address: "/s_new"}

  synth_id := int32(server.synth_ids)
  target_id := int32(0)
  add_action := int32(0)

  msg.Args = append(msg.Args, name, synth_id, add_action, target_id)
  if len(params) > 0 {
    msg.Args = append(msg.Args, params...)
  }

  server.synth_ids++;

  e := server.SendMessage(msg)

  return synth_id, e
}

func (server *Server) SetNodeControls (node_id int32, params ... interface {}) error {
  if len(params) % 2 != 0 {
    return sc_error { "Can not set node controls: odd number of parameters." }
  }

  msg := &osc.Message{Address: "/n_set"}
  msg.Args = append(msg.Args, node_id)
  msg.Args = append(msg.Args, params...)

  e := server.SendMessage(msg)
  return e
}

func (server *Server) FreeNode (id int32) error {
  msg := &osc.Message{Address: "/n_free"}
  msg.Args = append(msg.Args, id)
  e := server.SendMessage(msg)
  return e
}

