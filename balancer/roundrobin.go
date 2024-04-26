package balancer

import "sync"

var rrPtr = 0 // for selecting the next server in Round Robin fashion
func roundRobinInit(serverStream chan ServerConn, servers []server, mut *sync.RWMutex) {
	defer close(serverStream)
	for {
		serverStream <- ServerConn{servers[rrPtr].Address, servers[rrPtr].connEnd(), servers[rrPtr].connBegin()}
		rrPtr++
		if rrPtr%len(servers) == 0 {
			rrPtr = 0
		}
	}
}
