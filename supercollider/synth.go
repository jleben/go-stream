package supercollider

type Synth struct {
  valid bool
  server *Server
  id int32
}

func NewSynth ( server *Server, name string, params ... interface {} ) Synth {
  id, err := server.NewSynth(name, params...)
  valid := err == nil
  synth := Synth { valid, server, id }
  return synth
}

func (synth Synth) Set ( params ... interface {} ) {
  if !synth.valid { return }
  synth.server.SetNodeControls( synth.id, params... )
}

func (synth Synth) Free () {
  if !synth.valid { return }
  synth.server.FreeNode( synth.id )
  synth.valid = false
}
