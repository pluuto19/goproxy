package balancer

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

func healthCheckInit(servers []server, mut *sync.RWMutex) {
	for {
		for i := range servers {
			respBuffer := make([]byte, 1024)
			tcpAddr, err := net.ResolveTCPAddr("tcp4", servers[i].Address)
			if err != nil {
				fmt.Println(err)
				return
			}
			backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
			if err != nil {
				fmt.Println(err)
				return
			}

			reqMsg := "GET /health HTTP/1.1\r\nHost: " + strings.Split(servers[i].Address, ":")[0] + "\r\n"

			_, err1 := backendConnSock.Write([]byte(reqMsg))
			if err1 != nil {
				fmt.Println(err1)
				return
			}

			n, err := backendConnSock.Read(respBuffer)
			if err != nil {
				fmt.Println("Error reading response:", err)
				return
			}

			if !strings.Contains(string(respBuffer[:n]), "200") {
				mut.Lock()
				servers[i].Online = false
				mut.Unlock()
			} else {
				mut.RLock()
				isOnline := servers[i].Online
				mut.RUnlock()
				if !isOnline {
					mut.Lock() // expensive operation. better to check it first instead of blindly locking
					servers[i].Online = true
					mut.Unlock()
				}
			}
			time.Sleep(2 * time.Second)
			fmt.Println("Response from server:", string(respBuffer[:n]))
		}
	}
}

// implement in RR uses RLock
// implement in LC uses RLock
