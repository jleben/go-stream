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

func Iterate(list *list.List) chan stream.Event {
  
  work := func (output chan stream.Event) {
    for {
      for e := list.Front(); e != nil; e = e.Next() {
        output <- e.Value
      }
    }
  }
  
  return stream.Source(work)
}

func Compose( input ... chan stream.Event) chan stream.Event {
  output := make(chan stream.Event)

  work := func() {
    dur_in := input[0]
    text_in := input[1]
    for {
      dur := (<-dur_in).(int)
      text := (<-text_in).(string)
      e := Event {dur, text}
      output <- e
    }
  }

  go work()

  return output
}

func Play( source chan stream.Event, tatum time.Duration, reference time.Time ) chan stream.Event {
  output := make(chan stream.Event)

  work := func () {
    t := 0
    for token, ok := <-source; ok; token, ok = <-source {
      event := token.(Event)

      output <- event;

      t = t + event.Duration;

      real_time := reference.Add(time.Duration(t) * tatum);
      real_duration := real_time.Sub(time.Now());
      real_time = <-time.After(real_duration);
    }
  }

  go work()

  return output
}
