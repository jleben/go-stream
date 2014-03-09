/*
Task scheduler

Copyright (C) 2014 Jakob Leben <jakob.leben@gmail.com>

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

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
  Time () time.Time
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
  time time.Time
  queue pQueue
}

func NewScheduleStarting (start_time time.Time) *Schedule {
  return &Schedule{time: start_time}
}

func (this *Schedule) Time () time.Time { return this.time }

func (this *Schedule) SetTime( t time.Time ) { this.time = t }

func (this *Schedule) Schedule (task Task, delay time.Duration) {
  time := this.time.Add(delay)
  //fmt.Printf("Scheduling %v at %v\n", task, time)
  this.queue.Push( &pQueueItem{time, task} )
}

func (this *Schedule) Run () {

  for {
    now := time.Now()

    for this.queue.Len() > 0 && !this.queue.Top().time.After(now) {
      //fmt.Println("Will perform:", this.queue.Top())
      item := this.queue.Pop()
      this.time = item.time
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
