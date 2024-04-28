package balancer

import (
	"container/heap"
	"fmt"
	"sync"
)

type serverHeap []server

func (s serverHeap) Len() int           { return len(s) }
func (s serverHeap) Less(i, j int) bool { return s[i].ActiveConn < s[j].ActiveConn }
func (s serverHeap) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *serverHeap) Push(x any) {
	*s = append(*s, x.(server))
}

func (s *serverHeap) Pop() any {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

func leastConnInit(serverStream chan ServerConn, servers []server, onlineMut *sync.RWMutex) {
	// make a minheap of servers based on connection count
	// fill an unbuffered channel with instances fetched from heap in an inf for loop
	// check in connbeg and connend and if its LC dont increase and decrease once. instead increase once here.
	sHeap := make(serverHeap, len(servers))
	copy(sHeap, servers)
	heap.Init(&sHeap)
	for len(sHeap) > 0 {
		fmt.Println(heap.Pop(&sHeap))
	}
}

// what if physical system goes offline
