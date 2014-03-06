package main

import (
  "fmt"
  "time"
  //"./stream"
  "./muse3"
)

func main() {
  //tatum := 200 * time.Millisecond
  start_time := time.Now();

  /*
  play := func (dur []interface{}, text []interface{}) stream.Operator {
    //dur_stream := muse.Iterate( muse.ListInt(dur) )
    //text_stream := muse.Iterate( muse.ListString(text) )
    dur_stream := muse.Iterate( dur... )
    text_stream := muse.Iterate( text... )

    events := muse.Compose(dur_stream, "text", text_stream)

    return events
  }

  var voices [] stream.Operator;
  for i := 0; i < 10; i++ {
    voices = append(voices, play( []interface{}{1,3}, []interface{}{"a"} ) );
  }

  // Version 1:

  //x := muse.Play(tatum, start_time, voices...)

  // Version 2:

  for i := 0; i < 10; i++ {
    voices [i] = muse.PlayOne(voices[i], tatum, start_time)
  }
  x := stream.Join(voices...)

  //
  */

  x := muse.Repeat( muse.Iterate(1,2,3), 2 )

  s := x.Play()

  for {
    e, ok := <-s;
    if (!ok) { break }
    fmt.Printf("%v: %v\n", time.Now().Sub(start_time).Seconds() * 1000, e);
    time.Sleep(300 * time.Millisecond)
  }

}
