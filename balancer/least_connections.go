package balancer

import (
	"sync"
)

func leastConnInit(serverStream chan ServerConn, servers []server, onlineMut *sync.RWMutex) {
	// make a minheap of servers based on connection count
	// fill an unbuffered channel with instances fetched from heap in an inf for loop
	// check in connbeg and connend and if its LC dont increase and decrease once. instead increase once here.
}

// what if physical system goes offline
