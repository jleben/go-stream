package stream

type Event interface {}

//

type SourceFunc func (output chan Event)
type FilterFunc func (output chan Event, input ... chan Event )

func Source( play SourceFunc ) chan Event {
  output := make(chan Event)
  go play(output)
  return output
}

func Filter( play FilterFunc, input ... chan Event ) chan Event {
  output := make(chan Event)
  go play(output, input...)
  return output
}

//

func Split (input chan Event, count int) [] chan Event {
  var output [] chan Event

  for i := 0; i < count; i++ {
    output = append(output, make(chan Event))
  }

  work :=  func (input chan Event, output [] chan Event) {
    for event, ok := <- input; ok; event, ok = <- input {
      for i := 0; i < len(output); i++ {
        output[i] <- event
      }
    }
    for i := 0; i < len(output); i++ {
      close(output[i])
    }
  }

  go work(input, output)

  return output
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

func Join (inputs ... chan Event) chan Event {

  output := make(chan Event)

  work := func (input chan Event) {
    for {
        output <- <- input;
    }
  }

  for _, input := range inputs {
    go work(input)
  }

  return output
}
