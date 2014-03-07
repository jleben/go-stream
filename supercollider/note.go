package supercollider

import (
  "time"
  //"fmt"
  "../stream"
  "../muse3"
)

type note_end_event struct {
  time time.Time
  id int32
}

func (server *Server) await_note_end (request note_end_event, response chan note_end_event) {
  duration := request.time.Sub( time.Now() )
  time.Sleep(duration)
  response <- request
}

func (server *Server) perform_note_start( event muse.Event, response chan note_end_event ) bool {
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

  id, err := server.NewSynth(instrument, params...)
  if err == nil {
    real_duration := time.Duration(dur * 1000 * 1000) * time.Microsecond
    end_time := time.Now().Add(real_duration)
    go server.await_note_end( note_end_event{end_time, id}, response )
    return true
  } else {
    return false
  }

  //fmt.Printf("Note End: %v\n", node_id)
}

func (server *Server) perform_note_end( id int32 ) {
  server.SetNodeControls(id, "gate", float32(0))
  //fmt.Printf("Note Stop: %v\n", node_id)
}


func (server *Server) Play (music stream.Stream) {
  note_count := 0
  note_end := make(chan note_end_event)
  for {
    var ok bool
    var item stream.Item

    select {

      case item, ok = <-music:

        if ok {
          event := item.(muse.Event)
          if (server.perform_note_start(event, note_end)) {
            note_count++
          }
        } else {
          music = nil
        }

      case e := <-note_end:

        server.perform_note_end(e.id)
        note_count--

    }

    if music == nil && note_count == 0 { break }
  }
}
