package main

import (
  "fmt"
  "time"
  "math/rand"
  "math"
  //"./muse3"
  "./schedule"
  "./stream"
  "./supercollider"
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

  tatum := 50 * time.Millisecond
  start_time := time.Now()

  _ = tatum
  _ = start_time

  //server.DumpOSC(true)
  //server.DumpTree(0)

  //time.Sleep(300 * time.Millisecond)

  /*
  generator := func ( base_pitch float32 ) func(chan stream.Item) {
    f := func (output chan stream.Item) {
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
  }*/


  rand_pitch := func (output stream.Stream) {
    for {
      output <- int(rand.Float32() * 12)
    }
  }

  var _ = rand_pitch

  pitch_to_freq := func(output stream.Stream, inputs... stream.Stream) {
      pitch_in := inputs[0]
      for {
        pitch, ok := (<-pitch_in).(int)
        if !ok { break }
        freq := 440 * math.Pow(2, float64(pitch) / 12)
        output <- float32(freq)
      }
  }

  //pitch := stream.Series( stream.Series(1,2,3), stream.Series(6,8,4) )
  pitch := stream.Source(rand_pitch)

  score1 := supercollider.Compose( stream.Repeat(stream.Series(20,5,10), -1),
                                  "duration", stream.Const(float32(0.01)),
                                  "freq", stream.Filter(pitch_to_freq, pitch) )

  score2 := supercollider.Compose( stream.Repeat(stream.Series(20,20,30), -1),
                                  "duration", stream.Const(float32(0.01)),
                                  "freq", stream.Filter(pitch_to_freq, pitch) )
  //music := muse.Conduct(tatum, start_time, stream.Repeat(score, 2))


  /*
  music := muse.Conduct(tatum, start_time,
                     stream.Source(generator(600)),
                     stream.Source(generator(6000)) ).Play()
  */

  schedule := new(schedule.Schedule)
  schedule.SetTime( time.Now() )

  conductor := supercollider.NewConductor( server, schedule )
  conductor.Play(score1, score2)

  schedule.Run()


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
