package muse

import (
  "time"
  "../stream"
  "container/list"
)

type Event struct {
  Duration int
  Text string
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

func Compose( sources ... stream.Operator ) stream.Operator {
  work := func(output chan stream.Event, inputs... chan stream.Event) {
    dur_in := inputs[0]
    text_in := inputs[1]
    for {
      dur := (<-dur_in).(int)
      text := (<-text_in).(string)
      e := Event {dur, text}
      output <- e
    }
  }
  return stream.Filter(work, sources...)
}

func Play( source stream.Operator, tatum time.Duration, reference time.Time ) stream.Operator {

  work := func (output chan stream.Event, inputs... chan stream.Event) {
    input := inputs[0]
    t := 0
    for token, ok := <-input; ok; token, ok = <-input {
      event := token.(Event)

      output <- event;

      t = t + event.Duration;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      real_time = <-time.After(real_duration);
    }
  }

  return stream.Filter(work, source)
}
