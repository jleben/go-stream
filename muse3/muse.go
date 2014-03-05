package muse

import (
  "../stream"
  "container/list"
)

func ListInt( slice []int ) *list.List {
  list := list.New()
  for i := 0; i < len(slice); i++ {
    list.PushBack(slice[i])
  }
  return list;
}

func ListString( slice []string ) *list.List {
  list := list.New()
  for i := 0; i < len(slice); i++ {
    list.PushBack(slice[i])
  }
  return list;
}

func Iterate(list *list.List) chan stream.Event {
  
  work := func (output chan stream.Event) {
    for {
      for e := list.Front(); e != nil; e = e.Next() {
        output <- e.Value
      }
    }
  }
  
  return stream.Source(work)
}
