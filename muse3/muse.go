package muse

import (
  "time"
  "../stream"
  "container/heap"
)

type Event struct {
  Duration int
  Parameters map[string]interface{}
}

//

func Compose( duration stream.Operator, parameters ... interface {} ) stream.Operator {
  if len(parameters) % 2 != 0 {
    return nil
  }
  var keys [] string
  var sources [] stream.Operator
  sources = append(sources, duration)
  for i := 0; i < len(parameters); i = i + 2 {
    keys = append(keys, parameters[i].(string))
    sources = append(sources, parameters[i+1].(stream.Operator))
  }

  work := func(output stream.Stream, inputs... stream.Stream) {
    duration := inputs[0]
    parameters := inputs[1:]
    for {
      var ok bool

      e := Event{}
      e.Parameters = make( map[string]interface{} )

      var d stream.Item
      d, ok = <-duration
      if !ok { break }
      e.Duration = d.(int)

      for i, key := range keys {
        var p stream.Item
        p, ok = <-parameters[i]
        if !ok { break }
        e.Parameters[key] = p
      }

      if !ok { break }

      output <- e
    }
  }

  return stream.Filter(work, sources...)
}

func Conduct( tatum time.Duration, reference time.Time, sources... stream.Operator ) stream.Operator {

  work := func (output stream.Stream, inputs... stream.Stream) {

    q := &EventQueue{}
    t := 0

    for _, input := range inputs {
      stream := new(EventQueueItem)
      stream.source = input
      stream.time = 0
      heap.Push(q, stream)
    }

    for {

      for q.Len() > 0 && q.Top().time <= t {
        stream := heap.Pop(q).(*EventQueueItem)
        item, ok := <-stream.source
        if ok {
          event := item.(Event)

          output <- event

          stream.time = stream.time + event.Duration
          heap.Push(q, stream)
        }
      }

      if q.Len() == 0 {
        break
      }

      t = q.Top().time;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      time.Sleep(real_duration)
    }
  }

  return stream.Filter(work, sources...)
}


func ConductOne( source stream.Operator, tatum time.Duration, reference time.Time ) stream.Operator {

  work := func (output stream.Stream, inputs... stream.Stream) {
    input := inputs[0]
    t := 0
    for item, ok := <-input; ok; item, ok = <-input {
      event := item.(Event)

      output <- event;

      t = t + event.Duration;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      time.Sleep(real_duration)
    }
  }

  return stream.Filter(work, source)
}
