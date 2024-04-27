package balancer

import "sync"

var rrPtr = 0 // for selecting the next server in Round Robin fashion

func roundRobinInit(serverStream chan ServerConn, servers []server, mut *sync.RWMutex) {
	defer close(serverStream)
	for {
		mut.RLock()
		isServerOnline := servers[rrPtr].Online
		mut.RUnlock()
		if isServerOnline {
			serverStream <- ServerConn{servers[rrPtr].Address, servers[rrPtr].connEnd(), servers[rrPtr].connBegin(), &isServerOnline}
		}
		rrPtr++
		if rrPtr%len(servers) == 0 {
			rrPtr = 0
		}
	}
}

// refactor and remove ServerConn, use server struct as a channel type
