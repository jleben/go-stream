package main

import (
  "fmt"
  "time"
  "./schedule"
)

type ScheduleFunc func (schedule.Scheduler) time.Duration

func (f ScheduleFunc) Perform (s schedule.Scheduler) {
  delay := f(s)
  s.Schedule(f, delay)
}

type Tester struct {
  pause time.Duration
  latencies [] time.Duration
  index int
}

func (task *Tester) Perform (sched schedule.Scheduler) {
  if (task.index >= len(task.latencies)) { return }

  index := task.index
  task.index++;

  late := time.Now().Sub(sched.Time())
  task.latencies[index] = late

  sched.Schedule(task, task.pause)
}


func main() {
  /*
  sched := schedule.NewScheduleStarting( time.Now() )
  task := func (sched schedule.Scheduler) time.Duration {
    d := time.Now().Sub(sched.Time())
    fmt.Println("Latency = ", d)
    return 200 * time.Millisecond
  }

  sched.Schedule(ScheduleFunc(task), 0)
  */

  measurement_count := int(2e2)
  warm_up_count := int(30)
  inter_event_time := 30 * time.Millisecond

  tester := & Tester {
    inter_event_time,
    make([] time.Duration, measurement_count),
    0 }

  sched := schedule.NewScheduleStarting( time.Now() )
  sched.Schedule(tester, 0);
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
