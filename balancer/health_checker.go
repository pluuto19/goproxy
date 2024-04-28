package balancer

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var hcPtr = 0

func healthCheckInit(serversSlice []server, onlineMut *sync.RWMutex) {
	fmt.Println("Initial Health Check starting ...")
	for i := range serversSlice {
		respBuffer := make([]byte, 1024)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", serversSlice[i].Address)
		if err != nil {
			fmt.Println(err)
			return
		}
		backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Sent to " + serversSlice[i].Address)
		reqMsg := "GET /health HTTP/1.1\r\n" + "Host: " + strings.Split(serversSlice[i].Address, ":")[0] + "\r\n" + "Connection: close\r\n" + "User-Agent: Go-Health-Checker\r\n" + "\r\n"

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
		fmt.Println(string(respBuffer[:n]))
		if strings.Contains(string(respBuffer[:n]), "200") {
			onlineMut.Lock()
			serversSlice[i].Online = true
			onlineMut.Unlock()
		}
	}
	fmt.Println("Beginning Default Health Checks")
	go healthCheck(serversSlice, onlineMut)
}

func healthCheck(serversSlice []server, onlineMut *sync.RWMutex) {
	for {
		//fmt.Println("Health Checking " + serversSlice[hcPtr].Address)

		respBuffer := make([]byte, 1024)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", serversSlice[hcPtr].Address)
		if err != nil {
			fmt.Println(err)
			return
		}
		backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			fmt.Println(err)
			return
		}

		reqMsg := "GET /health HTTP/1.1\r\n" + "Host: " + strings.Split(serversSlice[hcPtr].Address, ":")[0] + "\r\n" + "Connection: close\r\n" + "User-Agent: Go-Health-Checker\r\n" + "\r\n"

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
			onlineMut.RLock()
			isOnline := serversSlice[hcPtr].Online
			onlineMut.RUnlock()
			if isOnline { // expensive operation. better to check it first instead of blindly locking
				onlineMut.Lock()
				serversSlice[hcPtr].Online = false
				onlineMut.Unlock()
			}
		} else {
			onlineMut.RLock()
			isOnline := serversSlice[hcPtr].Online
			onlineMut.RUnlock()
			if !isOnline { // expensive operation. better to check it first instead of blindly locking
				onlineMut.Lock()
				serversSlice[hcPtr].Online = true
				onlineMut.Unlock()
			}
		}
		hcPtr++
		if hcPtr%len(serversSlice) == 0 {
			hcPtr = 0
		}
		//fmt.Println("Response from server:", string(respBuffer[:n]))
		time.Sleep(1 * time.Second)
	}
}
