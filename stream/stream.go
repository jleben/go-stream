package stream

type Item interface {}

type Stream (chan Item)

type Operator interface {
  Stream () Stream
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

func (s *source) Stream () Stream {
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

func (f *filter) Stream () Stream {
  output := make(Stream)
  var inputs [] Stream
  for _, source := range f.sources {
    inputs = append(inputs, source.Stream())
  }
  work := func() {
    f.work(output, inputs...)
    close(output)
  }
  go work()
  return output
}

//

func Const(value interface {}) Operator {

  work := func (output Stream) {
    for { output <- value }
  }

  return Source(work)
}

//

func Series(items... interface{}) Operator {
  var inputs [] interface {}

  for _, item := range items {
    switch item := item.(type) {
      case Operator:
        inputs = append(inputs, item.Stream())
      default:
        inputs = append(inputs, item)
    }
  }

  work := func (output Stream) {
    for _, in := range inputs {
      switch input := in.(type) {
        case Stream: {
          for {
            value, ok := <-input
            if (ok) { output <- value } else { break }
          }
        }
        default:
          output <- input
      }
    }
  }

  return Source(work)
}

//

func Repeat(op Operator, times int) Operator {
  work := func (output Stream) {
    if times >= 0 {
      for i := 0; i < times; i++ {
        input := op.Stream()
        for {
          item, ok := <-input
          if ok { output <- item } else { break }
        }
      }
    } else {
      for {
        input := op.Stream()
        for {
          item, ok := <-input
          if ok { output <- item } else { break }
        }
      }
    }
  }
  return Source(work)
}

/*
type splitter_channel {
  output Stream
}

func Split (source Operator, count int) [] Operator {
  var channels [] Operator

  for i := 0; i < count; i++ {
    channels = append( channels, splitter_channel{ make(Stream) } )
  }

  input := source.Stream();

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
