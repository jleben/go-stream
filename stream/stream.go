package stream

type Event interface {}

type Operator interface {
  Play () chan Event
}

//

type SourceFunc func (output chan Event)

type source struct {
  work SourceFunc
}

func Source ( work SourceFunc ) Operator {
  s := new(source);
  s.work = work;
  return s
}

func (s *source) Play () chan Event {
  output := make(chan Event)
  work := func() {
    s.work(output)
    close(output)
  }
  go work()
  return output
}

//

type FilterFunc func (output chan Event, input ... chan Event )

type filter struct {
  work FilterFunc
  sources [] Operator
}

func Filter ( work FilterFunc, sources ... Operator ) Operator {
  f := new(filter)
  f.work = work
  f.sources = sources;
  return f
}

func (f *filter) Play () chan Event {
  output := make(chan Event)
  var inputs [] chan Event
  for _, source := range f.sources {
    inputs = append(inputs, source.Play())
  }
  work := func() {
    f.work(output, inputs...)
    close(output)
  }
  go work()
  return output
}

//

/*
type splitter_channel {
  output chan Event
}

func Split (source Operator, count int) [] Operator {
  var channels [] Operator

  for i := 0; i < count; i++ {
    channels = append( channels, splitter_channel{ make(chan Event) } )
  }

  input := source.Play();

  work :=  func () {
    for event, ok := <- input; ok; event, ok = <- input {
      for _, channel := range channels {
        channel.output <- event
      }
    }
    for _, channel := range channels {
        close(channel.output)
    }
  }

  go work()

  return channels
}

func Merge (inputs ... chan Event) chan Event {

  work := func (output chan Event, inputs ... chan Event) {
    for {
      for _, input := range inputs {
        output <- <- input;
      }
    }
  }

  return Filter(work, inputs...)
}
*/

func Join (sources ... Operator) Operator {
  forward := func(input chan Event, output chan Event) {
    for {
      output <- <- input
    }
  }

  work := func(output chan Event, inputs ... chan Event) {
    for _, input := range inputs {
      go forward(input, output)
    }
  }

  return Filter(work, sources...)
}

