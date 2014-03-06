package stream

type Item interface {}

type Stream (chan Item)

type Operator interface {
  Play () Stream
}


//

type SourceFunc func (output Stream)

type source struct {
  work SourceFunc
}

func Source ( work SourceFunc ) Operator {
  s := new(source);
  s.work = work;
  return s
}

func (s *source) Play () Stream {
  output := make(Stream)
  work := func() {
    s.work(output)
    close(output)
  }
  go work()
  return output
}

//

type FilterFunc func (output Stream, input ... Stream )

type filter struct {
  work FilterFunc
  sources [] Operator
}

func Filter ( work FilterFunc, sources ... Operator ) Operator {
  f := new(filter)
  f.work = work
  f.sources = sources
  return f
}

func (f *filter) Play () Stream {
  output := make(Stream)
  var inputs [] Stream
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
  output Stream
}

func Split (source Operator, count int) [] Operator {
  var channels [] Operator

  for i := 0; i < count; i++ {
    channels = append( channels, splitter_channel{ make(Stream) } )
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

func Merge (inputs ... Stream) Stream {

  work := func (output Stream, inputs ... Stream) {
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
  source_done := make(chan bool)
  all_done := make(chan bool)

  forward := func(input Stream, output Stream) {
    for {
      item, ok := <- input
      if ok {
        output <- item
      } else {
        source_done <- true
        break
      }
    }
  }

  cleanup := func(output Stream) {
    for i := 0; i < len(sources); i++ {
      <- source_done
    }
    all_done <- true
  }

  work := func(output Stream, inputs ... Stream) {
    for _, input := range inputs {
      go forward(input, output)
    }
    go cleanup(output)
    <- all_done
  }

  return Filter(work, sources...)
}

