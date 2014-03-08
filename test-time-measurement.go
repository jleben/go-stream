package main

import (
  "fmt"
  "time"
)

func main() {

  const count = 100
  t := make([]time.Time, count)

  for i := 0; i < count; i++ {
    t[i] = time.Now()
  }

  for i := 1; i < count; i++ {
    d := t[i].Sub(t[i-1])
    fmt.Printf("Difference = %v\n", d)
  }
}
