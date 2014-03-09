package stream

import ("fmt")

type Item interface {}

type Status int

const (
  Ok Status = iota;
  Closed
  Interrupted
)

// NOTE:

// A Pull will never be Interrupted.
// It is assumed that Pull never depends on downstream.

// A Push will never be Closed.
// Only writers ever close channels.

type Writer interface {
  Push (Item) Status
  Close ()
}

type Reader interface {
  Pull () (Item, Status)
  Close ()
}

type Operator interface {
  Stream() Reader
}

//

type Stream struct {
  data chan Item
  finish chan struct {}
}

func NewStream () *Stream {
  return &Stream { make(chan Item), make(chan struct{}) }
}

type StreamWriter Stream
type StreamReader Stream

func (s *StreamWriter) Push (data Item) Status {
  select {
    case s.data <- data:
      return Ok
    case <- s.finish:
      return Interrupted
  }
}

func (s *StreamWriter) Close () {
  close(s.data)
  <-s.finish
}

func (s *StreamReader) Pull () (Item, Status) {
  item, ok := <-s.data
  if ok {
    return item, Ok
  } else {
    return nil, Closed
  }
}

func (s *StreamReader) Close () {
  s.finish <- (struct{}{})
  close(s.finish)
}

//

type SourceFunc func ( Writer )

type SourceOp struct {
  work SourceFunc
}

func (op *SourceOp) Stream () Reader {
  output := NewStream()
  out_writer := (*StreamWriter)(output)
  out_reader := (*StreamReader)(output)

  work := func () {
    op.work( out_writer )
    out_writer.Close()
    fmt.Println("Source: finished")
  }
  go work()

  return out_reader
}

func Source ( work SourceFunc ) Operator {
  return & SourceOp { work }
}

//

type FilterFunc func ( Writer, ...Reader )

type FilterOp struct {
  work FilterFunc
  sources [] Operator
}

func (op *FilterOp) Stream () Reader {
  output := NewStream()
  out_writer := (*StreamWriter)(output)
  out_reader := (*StreamReader)(output)

  inputs := make([]Reader, len(op.sources))
  for i, source := range op.sources {
    inputs[i] = source.Stream()
  }

  work := func () {
    op.work( out_writer, inputs... )
    for _, input := range inputs { input.Close() }
    out_writer.Close()
    fmt.Println("Filter: finished")
  }
  go work()

  return out_reader
}

func Filter ( work FilterFunc, sources... Operator ) Operator {
  return & FilterOp { work, sources }
}
