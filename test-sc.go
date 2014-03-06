package main

import (
  "fmt"
  "time"
  "math/rand"
  "./muse3"
  "./supercollider"
  "./stream"
)

func report_error( e error ) {
  if e != nil { fmt.Println("Error:", e.Error()) }
}

func main() {
  var err error

  server, err := supercollider.NewServer("localhost:57110")
  if err != nil { fmt.Println("Error:", err.Error()) }

  err = server.Connect()
  if err != nil {
    fmt.Println("Error:", err.Error())
  } else {
    fmt.Println("Success!")
  }

  //server.DumpOSC(true)
  //server.DumpTree(0)

  //time.Sleep(300 * time.Millisecond)

  generator := func ( base_pitch float32 ) func(chan stream.Event) {
    f := func (output chan stream.Event) {
        e := muse.Event {}
        e.Duration = 10
        e.Parameters = map[string]interface{} {
          "type": "note-start",
          "instrument": "noise",
          "amp": float32(1),
          "duration": float32(0.01),
          "dur": float32(0.01),
          "decay": float32(0.01),
          "release": float32(0.01),
        }

        for {
          //e.Parameters["amp"] = rand.Float32()
          e.Parameters["freq"] = rand.Float32() * 1000 + base_pitch
          output <- e
        }
    }
    return f
  }

  tatum := 10 * time.Millisecond
  start_time := time.Now()
  music := muse.Play(tatum, start_time,
                     stream.Source(generator(600)),
                     stream.Source(generator(6000)) ).Play()

  server.Play(music)

/*
  fmt.Println("Bing")

  //id, err := server.NewSynth("default")
  //report_error(err)

  synth := supercollider.NewSynth(server, "default")

  time.Sleep(200 * time.Millisecond)

  synth.Stop()

  fmt.Println("Bong")

  //time.Sleep(500 * time.Millisecond)

  server.DumpTree(0)
*/
  time.Sleep(500 * time.Millisecond)

  server.Disconnect()
}
