package main

import (
  "fmt"
  "time"
  "./stream"
  "./muse3"
)

func main() {

  play := func (dur []int, text []string) stream.Operator {
    dur_stream := muse.Iterate( muse.ListInt(dur) )
    text_stream := muse.Iterate( muse.ListString(text) )
    //f := stream.Fork(x, 2)
    //z := stream.Join( f[0], f[1] )

    events := muse.Compose(dur_stream, "text", text_stream)

    return events
  }

  var voices [] stream.Operator;
  for i := 0; i < 10; i++ {
    voices = append(voices, play( []int{1,3}, []string{"a"} ) );
  }

  tatum := 200 * time.Millisecond
  start_time := time.Now();

  // Version 1:

  //x := muse.Play(tatum, start_time, voices...)

  // Version 2:

  for i := 0; i < 10; i++ {
    voices [i] = muse.PlayOne(voices[i], tatum, start_time)
  }
  x := stream.Join(voices...)

  //

  s := x.Play()

  for {
    e := (<-s).(muse.Event);
    fmt.Printf("%v: %v\n", time.Now().Sub(start_time).Seconds() * 1000, e);
  }

}
