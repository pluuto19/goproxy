package balancer

import (
	"container/heap"
	"fmt"
	"sync"
)

var unbufServerStream chan ServerConn

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

func leastConnInit(servers []server, activeConn *sync.Mutex, onlineMut *sync.RWMutex) chan ServerConn {
	// check in connbeg and connend and if its LC dont increase and decrease once. instead increase once here.
	unbufServerStream = make(chan ServerConn)

	sHeap := make(serverHeap, len(servers))
	copy(sHeap, servers)
	heap.Init(&sHeap)

	//for len(sHeap) > 0 {
	//	fmt.Println(heap.Pop(&sHeap))
	//}
	//
	go LCPopulateChannel(sHeap, activeConn, onlineMut)
	return unbufServerStream
}

func LCPopulateChannel(sHeap serverHeap, mut *sync.Mutex, onlineMut *sync.RWMutex) {
	for {
		// peek the minimum then increase its active conn then fix it
		onlineMut.RLock()
		isServerOnline := sHeap[0].Online
		onlineMut.RUnlock()
		if isServerOnline {
			fmt.Println("heap before")
			fmt.Println(sHeap)
			fmt.Println("adding " + sHeap[0].Address)
			mut.Lock()
			sHeap[0].ActiveConn++
			mut.Unlock()
			// Call connBegin with the appropriate method
			unbufServerStream <- ServerConn{
				ServerAddr: sHeap[0].Address,
				ConnEnd:    sHeap[0].connEnd(),
				ConnBegin:  sHeap[0].connBegin(),
				IsOnline:   &sHeap[0].Online,
			}
		}
		heap.Fix(&sHeap, 0)
		fmt.Println("heap after")
		fmt.Println(sHeap)
	}
}

// what if physical system goes offline
// initial health check fail
// system behavior if a server goes down while load balancing
