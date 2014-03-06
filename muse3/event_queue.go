// Based on example of priority queue at: http://golang.org/pkg/container/heap/

package muse

import (
  "../stream"
)

// An Item is something we manage in a priority queue.
type EventQueueItem struct {
  source chan stream.Item
  time int    // The priority of the item in the queue.
  // The index is needed by update and is maintained by the heap.Interface methods.
  //index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type EventQueue []*EventQueueItem

func (q EventQueue) Len() int { return len(q) }

func (q EventQueue) Less(i, j int) bool {
  // We want Pop to give us the highest, not lowest, priority so we use greater than here.
  return q[i].time < q[j].time
}

func (q EventQueue) Swap(i, j int) {
  q[i], q[j] = q[j], q[i]
}

func (q *EventQueue) Push(x interface{}) {
  item := x.(*EventQueueItem)
  *q = append(*q, item)
}

func (q *EventQueue) Pop() interface{} {
  old := *q
  n := len(old)
  item := old[n-1]
  *q = old[0 : n-1]
  return item
}
