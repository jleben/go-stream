package supercollider

import (
  "time"
  "fmt"
  "../osc"
  "../stream"
  "../priority_queue"
  "../schedule"
)

var _ stream.Stream

// Musical stream interface

type EventParameters map[string]interface{}

type Event struct {
  Delay time.Duration
  Parameters EventParameters
}

// Compose...

type Composer struct {
  Duration stream.Operator
  Parameters (map [string] stream.Operator)
}

func Compose( duration stream.Operator, parameters ... interface {} ) *Composer {
  if len(parameters) % 2 != 0 {
    return nil
  }

  composer := & Composer { duration, make(map [string] stream.Operator) }

  for i := 0; i < len(parameters); i = i + 2 {
    key := parameters[i].(string)
    value := parameters[i+1].(stream.Operator)
    composer.Parameters[key] = value
  }

  return composer
}

func (this *Composer) Stream() stream.Reader {
  output := stream.NewStream()
  output_writer := (*stream.StreamWriter)(output)
  output_reader := (*stream.StreamReader)(output)

  duration := this.Duration.Stream()

  parameters := make(map [string] stream.Reader)
  for key, value := range this.Parameters {
    parameters[key] = value.Stream()
  }

  work := func() {

    InputProcessing:
    for {
      var status stream.Status

      event := Event{}
      event.Parameters = make(EventParameters)

      var d stream.Item
      d, status = duration.Pull()
      if status != stream.Ok { break }
      event.Delay = time.Duration(d.(int)) * 10 * time.Millisecond

      for key, value := range parameters {
        var p stream.Item
        p, status = value.Pull()
        if status != stream.Ok { break InputProcessing }
        event.Parameters[key] = p
      }

      status = output_writer.Push(event)
      if status == stream.Interrupted { break }
    }

    duration.Close()
    for _, input := range parameters { input.Close() }
    output_writer.Close()

    fmt.Println("Composer finished.")
  }

  go work()

  return output_reader
}

// Conduct...

// Conductor queue items

type queue_item struct {
  time time.Time
  task interface {}
}

func (this *queue_item) Less (other priority_queue.Item) bool {
  return this.time.Before( other.(*queue_item).time )
}

// Conductor queued tasks

type NoteEnd struct {
  id int32
}

type NoteProvider struct {
  stream stream.Reader
}

// Conductor

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

func (this *Conductor) Play( operators ... stream.Operator ) {

  if !this.scheduled { this.time = this.scheduler.Time() }

  for _, op := range operators {
    this.queue.Push( &queue_item { this.time, NoteProvider{op.Stream()} } )
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
        data, status := task.stream.Pull()
        if status != stream.Ok { break }
        event := data.(Event)
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
