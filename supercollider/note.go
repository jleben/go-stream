package supercollider

import (
  "time"
  "fmt"
  "../osc"
  "../stream"
  "../muse3"
  "../priority_queue"
  "../schedule"
)

var _ stream.Stream
var _ muse.Event

// Musical stream interface

type EventParameters map[string]interface{}

type Event struct {
  Delay time.Duration
  Parameters EventParameters
}

type Stream chan Event

//

/*
type StreamPlayer struct {
  server *Server
  stream Stream
}

type NoteEnd struct {
  server *Server
  id int32
}

func (this Stream) Perform ( scheduler schedule.Scheduler ) {
  event, ok := <-this
  if (!ok) { return }
  fmt.Println("Event:", event)
  scheduler.Schedule(this, event.Delay)
}

func (this *StreamPlayer) Perform ( scheduler schedule.Scheduler ) {
  event, ok := <-this.stream
  if (!ok) { return }

  fmt.Println("Note start:", event)

  dur := float32(1)
  instrument := "default"

  var params [] interface {}
  for key, value := range event.Parameters {
    switch key {
      //case "type":;
      case "duration":
        dur = value.(float32)
        params = append(params, key, value)
      case "instrument":
        instrument = value.(string)
      default:
        params = append(params, key, value)
    }
  }

  id, err := this.server.NewSynth(instrument, params...)
  if err == nil {
    note_end_delay := time.Duration(dur * 1000 * 1000) * time.Microsecond
    note_end := &NoteEnd {this.server, id}
    scheduler.Schedule(note_end, note_end_delay)
  }

  scheduler.Schedule(this, event.Delay)
}

func (this *NoteEnd) Perform ( sched schedule.Scheduler ) {
  fmt.Println("Node end")
  this.server.SetNodeControls(this.id, "gate", float32(0))
}
*/

// Queue items

type queue_item struct {
  time time.Time
  task interface {}
}

type NoteEnd struct {
  id int32
}

type NoteProvider struct {
  stream Stream
}

func (this *queue_item) Less (other priority_queue.Item) bool {
  return this.time.Before( other.(*queue_item).time )
}

//

type Conductor struct {
  server *Server
  scheduler schedule.Scheduler
  scheduled bool
  time time.Time
  queue priority_queue.Queue
}

func NewConductor (server *Server, scheduler schedule.Scheduler) *Conductor {
  //this := &Conductor{server, scheduler, false, _, _}
  this := new(Conductor)
  this.server = server
  this.scheduler = scheduler
  this.scheduled = false
  return this
}

func (this *Conductor) Time () time.Time { return this.time }

/*
func (this *Conductor) Schedule (item *queue_item, after time.Duration) {
  this.queue.Push(
}
*/

func (this *Conductor) Play( streams ... Stream ) {

  if !this.scheduled { this.time = this.scheduler.Time() }

  for _, stream := range streams {
    this.queue.Push( &queue_item { this.time, NoteProvider{stream} } )
  }

  if !this.scheduled { this.scheduler.Schedule( this, 0 ) }
}

func (this *Conductor) Perform( scheduler schedule.Scheduler ) {

  now := scheduler.Time()
  bundle := &osc.Bundle{Timetag: now}

  fmt.Println("S:", now)

  for this.queue.Len() > 0 && !this.queue.Top().(*queue_item).time.After(now) {
    //fmt.Println("Will perform:", this.queue.Top())
    item := this.queue.Pop().(*queue_item)
    this.time = item.time
    fmt.Println("P:", this.time)

    switch task := item.task.(type) {
      case NoteProvider:
        event, ok := <-task.stream
        if (!ok) { break }
        fmt.Println("Conductor: Note start!")
        item.time = this.time.Add( event.Delay )
        this.queue.Push( item )

        msg, id := this.note_start_message(event)
        if msg == nil { break }
        bundle.Messages = append(bundle.Messages, msg)

        event_dur := float32(1)
        dur_param := event.Parameters["duration"]
        if dur_param != nil { event_dur = dur_param.(float32) }
        note_dur := time.Duration(event_dur * 1000 * 1000) * time.Microsecond
        note_end_time := this.time.Add( note_dur )
        note_end := &queue_item{ note_end_time, NoteEnd{id} }
        this.queue.Push(note_end)

      case NoteEnd:
        fmt.Println("Conductor: Note end!")
        msg := this.note_end_message(task.id)
        bundle.Messages = append(bundle.Messages, msg)
    }
  }

  if len(bundle.Messages) > 0 {
    this.server.SendBundle(bundle)
  }

  if this.queue.Len() == 0 {
    return
  }

  next_time := this.queue.Top().(*queue_item).time;
  delay := next_time.Sub(now);
  scheduler.Schedule(this, delay)
}

func (this *Conductor) note_start_message ( event Event ) (*osc.Message, int32) {
  //dur := float32(1)
  instrument := "default"

  var params [] interface {}
  for key, value := range event.Parameters {
    switch key {
      //case "type":;
      case "duration":
        //dur = value.(float32)
        params = append(params, key, value)
      case "instrument":
        instrument = value.(string)
      default:
        params = append(params, key, value)
    }
  }

  return this.server.NewSynthMsg(instrument, params...)
}

func (this *Conductor) note_end_message (id int32) *osc.Message {
  return this.server.SetNodeControlsMsg(id, "gate", float32(0))
}

func Compose( duration_op stream.Operator, parameters ... interface {} ) Stream {
  if len(parameters) % 2 != 0 {
    return nil
  }

  duration := duration_op.Stream()

  var keys [] string
  var values [] stream.Stream
  for i := 0; i < len(parameters); i = i + 2 {
    keys = append(keys, parameters[i].(string))
    values = append(values, parameters[i+1].(stream.Operator).Stream())
  }

  output := make(Stream)

  work := func() {
    for {
      var ok bool

      e := Event{}
      e.Parameters = make(EventParameters)

      var d stream.Item
      d, ok = <-duration
      if !ok { break }
      e.Delay = time.Duration(d.(int)) * 10 * time.Millisecond

      for i, key := range keys {
        var p stream.Item
        p, ok = <-values[i]
        if !ok { break }
        e.Parameters[key] = p
      }

      if !ok { break }

      output <- e
    }

    close(output)
  }

  go work()

  return output
}
