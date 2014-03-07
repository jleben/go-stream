package main

import (
  "fmt"
  "time"
  "./stream"
  "./muse3"
)

func main() {
  tatum := 200 * time.Millisecond
  start_time := time.Now();

  fmt.Println("dummy", stream.Join())
  fmt.Println("dummy", muse.Compose(stream.Series()))
  _ = tatum
  _ = start_time

  /*
  x := muse.Compose( stream.Repeat(stream.Series(1,2), -2),
                     "amp:", stream.Repeat(stream.Series(0.5, 0.1), -1) )
  y := muse.Conduct(tatum, start_time, x)
  s := y.Stream()//stream.Join(x,y).Play()
  */

  x := stream.Repeat( muse.Compose( stream.Repeat(stream.Series(1,2),-1), "a", stream.Series(1,2,3) ), 2 )
  s := muse.Conduct(tatum, start_time, x).Stream()

  for {
    e, ok := <-s;
    if (!ok) { break }
    //fmt.Printf("%v: %v\n", time.Now().Sub(start_time).Seconds() * 1000, e);
    fmt.Println("<<", e)
    //time.Sleep(100 * time.Millisecond)
  }

}
