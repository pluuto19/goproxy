package balancer

import (
	"fmt"
	"sync"
)

var rrPtr = 0 // for selecting the next server in Round Robin fashion
var bufServerStream chan ServerConn

func roundRobinInit(servers []server, mut *sync.RWMutex) chan ServerConn {
	bufServerStream = make(chan ServerConn, 10000)
	go RRPopulateChannel(servers, mut)
	return bufServerStream
}

func RRPopulateChannel(servers []server, mut *sync.RWMutex) {
	defer close(bufServerStream)
	for {
		mut.RLock()
		isServerOnline := servers[rrPtr].Online
		mut.RUnlock()
		if isServerOnline {
			bufServerStream <- ServerConn{servers[rrPtr].Address, servers[rrPtr].connEnd(), servers[rrPtr].connBegin(), &isServerOnline}
		} else {
			fmt.Println(servers[rrPtr].Address + " was not online")
		}
		rrPtr++
		if rrPtr%len(servers) == 0 {
			rrPtr = 0
		}
	}
}
