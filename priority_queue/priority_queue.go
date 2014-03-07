// Based on example of priority queue at: http://golang.org/pkg/container/heap/

package priority_queue

import "container/heap"

type Item interface {
  Less(Item) bool
}

type storage []Item

func (this storage) Len() int { return len(this) }

func (this storage) Less(i, j int) bool {
  return this[i].Less(this[j])
}

func (this storage) Swap(i, j int) {
  this[i], this[j] = this[j], this[i]
}

func (this *storage) Push(item interface{}) {
  *this = append(*this, item.(Item))
}

func (this *storage) Pop() interface{} {
  old := *this
  n := len(old)
  item := old[n-1]
  *this = old[0 : n-1]
  return item
}

//

type Queue storage

func (this *Queue) Push(item Item) {
  heap.Push((*storage)(this), item)
}

func (this *Queue) Pop() Item {
  return heap.Pop((*storage)(this)).(Item)
}

func (this *Queue) At(index int) Item {
  return (*this)[index]
}

func (this *Queue) Top() Item {
  return (*this)[0]
}

func (this *Queue) Len() int {
  return (storage)(*this).Len()
}
