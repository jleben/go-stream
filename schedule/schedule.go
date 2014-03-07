package schedule

import (
  "time"
  "../priority_queue"
)

// Scheduling interface

type Task interface {
  Perform( Scheduler )
}

type Scheduler interface {
  Schedule ( task Task, delay time.Duration )
}

// Scheduling implementation

type pQueueItem struct {
  time time.Time
  task Task
}

func (this *pQueueItem) Less (other priority_queue.Item) bool {
  return this.time.Before( other.(*pQueueItem).time )
}

type pQueue struct {
  priority_queue.Queue
}

func (q *pQueue) Top () *pQueueItem {
  return q.Queue.Top().(*pQueueItem)
}

func (q *pQueue) Push( item *pQueueItem ) {
  q.Queue.Push(item)
}

func (q *pQueue) Pop () *pQueueItem {
  return q.Queue.Pop().(*pQueueItem)
}

type Schedule struct {
  Time time.Time
  queue pQueue
}

func (this *Schedule) Schedule (task Task, delay time.Duration) {
  time := this.Time.Add(delay)
  //fmt.Printf("Scheduling %v at %v\n", task, time)
  this.queue.Push( &pQueueItem{time, task} )
}

func (this *Schedule) Run () {

  for {
    now := time.Now()

    for this.queue.Len() > 0 && !this.queue.Top().time.After(now) {
      //fmt.Println("Will perform:", this.queue.Top())
      item := this.queue.Pop()
      this.Time = item.time
      item.task.Perform(this)
    }

    if this.queue.Len() == 0 {
      break
    }

    next_time := this.queue.Top().time;
    delay := next_time.Sub(time.Now());
    time.Sleep(delay)
  }

}
