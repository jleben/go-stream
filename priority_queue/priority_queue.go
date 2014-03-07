// Based on example of priority queue at: http://golang.org/pkg/container/heap/

package priority_queue

// An Item is something we manage in a priority queue.
type Item interface {
  Less(Item) bool
}

// A PriorityQueue implements heap.Interface and holds Items.
type Queue []Item

func (q Queue) Len() int { return len(q) }

func (q Queue) Less(i, j int) bool {
  // We want Pop to give us the highest, not lowest, priority so we use greater than here.
  return q[i].Less(q[j])
}

func (q Queue) Swap(i, j int) {
  q[i], q[j] = q[j], q[i]
}

func (q *Queue) Push(item interface{}) {
  *q = append(*q, item.(Item))
}

func (q *Queue) Pop() interface{} {
  old := *q
  n := len(old)
  item := old[n-1]
  *q = old[0 : n-1]
  return item
}

func (q *Queue) At(index int) Item {
  return (*q)[index]
}
