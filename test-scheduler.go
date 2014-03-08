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

func main() {

  sched := schedule.NewScheduleStarting( time.Now() )

  task := func (sched schedule.Scheduler) time.Duration {
    d := time.Now().Sub(sched.Time())
    fmt.Println("Latency = ", d)
    return 200 * time.Millisecond
  }

  sched.Schedule(ScheduleFunc(task), 0)

  sched.Run()
}
