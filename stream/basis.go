package stream

import ("fmt")

// Interface...

type Item interface {}

type Status int

const (
  Ok Status = iota;
  Closed
  Interrupted
)

type Writer interface {
  Push (Item) Status
  Close ()
}

type Reader interface {
  Pull () (Item, Status)
  Close ()
}

// NOTE:

// A Push will never return Closed.
// Only writers ever close channels.

// A Pull will never return Interrupted.
// It is assumed that Pull never depends on downstream, and hence would never
// block forever: it would either succeed or end because the upstream has ended
// and closed the channel.

type Operator interface {
  // An Operator represents a definition of a stream,
  // whereas Stream() instantiates a concrete stream of values
  // based on the definition.
  // This supports operator re-use:
  // streams can be instantiated from them multiple times
  Stream() Reader
}

// Implementation...

type Stream struct {
  // Transfer stream output downstream:
  data chan Item
  // Request from downstream for this stream to end:
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

// Generic source

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

// Generic filter

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
