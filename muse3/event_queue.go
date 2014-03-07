package muse

import (
  "../stream"
  "../priority_queue"
)

// An Item is something we manage in a priority queue.
type EventQueueItem struct {
  source chan stream.Item
  time int
}

func (this *EventQueueItem) Less (other priority_queue.Item) bool {
  return this.time < other.(*EventQueueItem).time
}

type EventQueue struct {
  priority_queue.Queue
}

func (q *EventQueue) Top () *EventQueueItem {
  return q.At(0).(*EventQueueItem)
}
