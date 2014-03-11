package main

import (
  "fmt"
  "time"
)

func main() {

  const measurement_count = int(1e5)
  const warm_up_count = int(100)

  t := make([]time.Time, measurement_count)

  for i := 0; i < measurement_count; i++ {
    t[i] = time.Now()
  }

  var min time.Duration = 1 * time.Second
  var max time.Duration = 0
  max_index := -1
  min_index := -1

  for i := warm_up_count; i < measurement_count; i++ {
    d := t[i].Sub(t[i-1])
    if d < min { min = d; min_index = i }
    if d > max { max = d; max_index = i }
  }

  fmt.Printf("Min: %v @ %v\n", min, min_index)
  fmt.Printf("Max: %v @ %v\n", max, max_index)
}
