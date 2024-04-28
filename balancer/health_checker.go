package balancer

import (
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"
)

var hcPtr = 0

func healthCheckInit(serversSlice []server, activeConnMut *sync.Mutex, onlineMut *sync.RWMutex, healthCheckWG *sync.WaitGroup) {
	fmt.Println("Initial Health Check starting ...")
	for i := range serversSlice {
		respBuffer := make([]byte, 1024)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", serversSlice[i].Address)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, i)
			fmt.Println(err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, i)
		}
		backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, i)
			fmt.Println(err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, i)
		}

		fmt.Println("Sent to " + serversSlice[i].Address)
		reqMsg := "GET /health HTTP/1.1\r\n" + "Host: " + strings.Split(serversSlice[i].Address, ":")[0] + "\r\n" + "Connection: close\r\n" + "User-Agent: Go-Health-Checker\r\n" + "\r\n"

		_, err1 := backendConnSock.Write([]byte(reqMsg))
		if err1 != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, i)
			fmt.Println(err1)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, i)
		}

		n, err := backendConnSock.Read(respBuffer)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, i)
			fmt.Println("Error reading response:", err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, i)
		}

		fmt.Println(string(respBuffer[:n]))
		if strings.Contains(string(respBuffer[:n]), "200") {
			onlineMut.Lock()
			serversSlice[i].Online = true
			onlineMut.Unlock()
		}
	}
	fmt.Println("Beginning Default Health Checks")
	healthCheckWG.Done()
	go healthCheck(serversSlice, activeConnMut, onlineMut)
}

func healthCheck(serversSlice []server, activeConnMut *sync.Mutex, onlineMut *sync.RWMutex) {
	for {
		respBuffer := make([]byte, 1024)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", serversSlice[hcPtr].Address)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, hcPtr)
			fmt.Println(err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, hcPtr)
		}
		backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, hcPtr)
			fmt.Println(err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, hcPtr)
		}

		reqMsg := "GET /health HTTP/1.1\r\n" + "Host: " + strings.Split(serversSlice[hcPtr].Address, ":")[0] + "\r\n" + "Connection: close\r\n" + "User-Agent: Go-Health-Checker\r\n" + "\r\n"

		_, err1 := backendConnSock.Write([]byte(reqMsg))
		if err1 != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, hcPtr)
			fmt.Println(err1)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, hcPtr)
		}

		n, err := backendConnSock.Read(respBuffer)
		if err != nil {
			putServerOffline(serversSlice, activeConnMut, onlineMut, hcPtr)
			fmt.Println("Error reading response:", err)
			return
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, hcPtr)
		}

		if !strings.Contains(string(respBuffer[:n]), "200") {
			putServerOffline(serversSlice, activeConnMut, onlineMut, hcPtr)
		} else {
			putServerOnline(serversSlice, activeConnMut, onlineMut, hcPtr)
		}

		hcPtr++
		if hcPtr%len(serversSlice) == 0 {
			hcPtr = 0
		}
		time.Sleep(1 * time.Second)
	}
}

func putServerOffline(serversSlice []server, activeConnMut *sync.Mutex, onlineMut *sync.RWMutex, index int) {
	onlineMut.RLock()
	isOnline := serversSlice[index].Online
	onlineMut.RUnlock()
	if isOnline { // expensive operation. better to check it first instead of blindly locking
		onlineMut.Lock()
		serversSlice[index].Online = false
		onlineMut.Unlock()
		activeConnMut.Lock()
		serversSlice[index].ActiveConn = math.MaxUint
		activeConnMut.Unlock()
	}
}
func putServerOnline(serversSlice []server, activeConnMut *sync.Mutex, onlineMut *sync.RWMutex, index int) {
	onlineMut.RLock()
	isOnline := serversSlice[index].Online
	onlineMut.RUnlock()
	if !isOnline { // expensive operation. better to check it first instead of blindly locking
		onlineMut.Lock()
		serversSlice[index].Online = true
		onlineMut.Unlock()
		activeConnMut.Lock()
		serversSlice[index].ActiveConn = 0
		activeConnMut.Unlock()
	}
}
