package muse

import (
  "time"
  "../stream"
  "container/list"
  "container/heap"
)

type Event struct {
  Duration int
  Parameters map[string]interface{}
}

func ListInt( slice []int ) *list.List {
  list := list.New()
  for i := 0; i < len(slice); i++ {
    list.PushBack(slice[i])
  }
  return list;
}

func ListString( slice []string ) *list.List {
  list := list.New()
  for i := 0; i < len(slice); i++ {
    list.PushBack(slice[i])
  }
  return list;
}

//

func Const(value interface {}) stream.Operator {

  work := func (output chan stream.Event) {
    for { output <- value }
  }

  return stream.Source(work)
}

func Iterate(list *list.List) stream.Operator {

  work := func (output chan stream.Event) {
    for {
      for e := list.Front(); e != nil; e = e.Next() {
        output <- e.Value
      }
    }
  }

  return stream.Source(work)
}

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

  work := func(output chan stream.Event, inputs... chan stream.Event) {
    duration := inputs[0]
    parameters := inputs[1:]
    for {
      e := Event{}
      e.Parameters = make( map[string]interface{} )
      e.Duration = (<-duration).(int)
      for i, key := range keys {
        e.Parameters[key] = <-parameters[i]
      }
      output <- e
    }
  }

  return stream.Filter(work, sources...)
}

func Play( tatum time.Duration, reference time.Time, sources... stream.Operator ) stream.Operator {

  work := func (output chan stream.Event, inputs... chan stream.Event) {

    q := &EventQueue{}
    t := 0

    for _, input := range inputs {
      stream := new(EventQueueItem)
      stream.source = input
      stream.time = 0
      heap.Push(q, stream)
    }

    for {

      for q.Len() > 0 && (*q)[0].time <= t {
        stream := heap.Pop(q).(*EventQueueItem)
        token, ok := <-stream.source
        if ok {
          event := token.(Event)

          output <- event

          stream.time = stream.time + event.Duration
          heap.Push(q, stream)
        }
      }

      if q.Len() == 0 {
        break
      }

      t = (*q)[0].time;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      time.Sleep(real_duration)
    }
  }

  return stream.Filter(work, sources...)
}


func PlayOne( source stream.Operator, tatum time.Duration, reference time.Time ) stream.Operator {

  work := func (output chan stream.Event, inputs... chan stream.Event) {
    input := inputs[0]
    t := 0
    for token, ok := <-input; ok; token, ok = <-input {
      event := token.(Event)

      output <- event;

      t = t + event.Duration;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      time.Sleep(real_duration)
    }
  }

  return stream.Filter(work, source)
}
