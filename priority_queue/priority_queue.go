/*
A simple, generic priority queue
Based on the example at: http://golang.org/pkg/container/heap/

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
