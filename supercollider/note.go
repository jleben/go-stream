package supercollider

import (
  "time"
  //"fmt"
  "../stream"
  "../muse3"
)

func pop (m map [string] interface {}, key string) interface{} {
  value := m[key]
  delete(m, key)
  return value
}

func (server *Server) perform_note_start( event muse.Event, comm stream.Stream ) {
  dur := float32(1)
  instrument := "default"

  var params [] interface {}
  for key, value := range event.Parameters {
    switch key {
      case "type":;
      case "duration": dur = value.(float32)
      case "instrument": instrument = value.(string)
      default: params = append(params, key, value)
    }
  }

  node_id, err := server.NewSynth(instrument, params...)
  if err == nil {
    cleanup := func () {
      time.Sleep(time.Duration(dur * 1000 * 1000) * time.Microsecond)
      e := muse.Event{}
      e.Parameters = map[string]interface{} {
        "type": "note-end",
        "node": node_id,
      }
      comm <- e
    }
    go cleanup()
  }

  //fmt.Printf("Note End: %v\n", node_id)
}

func (server *Server) perform_note_end( event muse.Event, stream stream.Stream ) {
  node_id := event.Parameters["node"].(int32)
  server.SetNodeControls(node_id, "gate", float32(0))

  //fmt.Printf("Note Stop: %v\n", node_id)
}

func (server *Server) Play (stream stream.Stream) {
  for {
    event := (<-stream).(muse.Event)
    event_type := event.Parameters["type"].(string)
    switch event_type {
      case "note-start": server.perform_note_start(event, stream);
      case "note-end": server.perform_note_end(event, stream);
    }
  }
}
