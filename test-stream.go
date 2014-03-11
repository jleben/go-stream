package main

import (
  "fmt"
  "time"
  . "./stream"
  "./schedule"
)

type StreamTester struct {
  stream Reader
  pause time.Duration
  latencies [] time.Duration
  index int
}

func (task *StreamTester) Perform (sched schedule.Scheduler) {
  if (task.index >= len(task.latencies)) { return }

  index := task.index
  task.index++;

  output, status := task.stream.Pull()
  _ = output
  if status == Ok {
    late := time.Now().Sub(sched.Time())
    //fmt.Printf("Late = %v, Value = %v\n", late, output)
    task.latencies[index] = late
    sched.Schedule(task, task.pause)
  }
}


func main() {

  a := Series( Repeat( Series(1,2,3), 2 ),
               Series( Series(5,6), Series(7,8), Series(9,10)  ) )

  r := Repeat( a, -1 )

  measurement_count := int(2e2)
  warm_up_count := int(30)
  inter_event_time := 30 * time.Millisecond

  sched := schedule.NewScheduleStarting( time.Now() )

  fmt.Println("Test one:")
  {
    tester := &StreamTester{}
    tester.stream = r.Stream()
    tester.pause = inter_event_time
    tester.latencies = make([]time.Duration, measurement_count)
    tester.index = 0

    sched.Schedule(tester, 0)
    sched.Run()

    var min time.Duration = 1 * time.Second
    var max time.Duration = 0
    max_index := -1
    min_index := -1

    for i := warm_up_count; i < len(tester.latencies); i++ {
      latency := tester.latencies[i]
      if latency < min { min = latency; min_index = i }
      if latency > max { max = latency; max_index = i }
    }

    fmt.Printf("Min: %v @ %v\n", min, min_index)
    fmt.Printf("Max: %v @ %v\n", max, max_index)

  }

  fmt.Println("Test many:")
  {
    sched.SetTime(time.Now())

    testers := make([]*StreamTester, 10);

    for i := 0; i < len(testers); i++ {
      tester := &StreamTester{}
      tester.stream = r.Stream()
      tester.pause = inter_event_time
      tester.latencies = make([]time.Duration, measurement_count)
      tester.index = 0

      testers[i] = tester

      sched.Schedule(tester, 100 * time.Millisecond)
    }

    sched.Run()

    var min time.Duration = 1 * time.Second
    var max time.Duration = 0
    max_index := -1
    min_index := -1

    for _, tester := range testers {
      for i := warm_up_count; i < len(tester.latencies); i++ {
        latency := tester.latencies[i]
        if latency < min { min = latency; min_index = i }
        if latency > max { max = latency; max_index = i }
      }
    }

    fmt.Printf("Min: %v @ %v\n", min, min_index)
    fmt.Printf("Max: %v @ %v\n", max, max_index)
  }
}
