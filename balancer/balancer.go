package balancer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const RR = 1
const LC = 2

var servers []server             // contains all server addresses from the JSON and their current active connection/s
var serverStream chan ServerConn // for sharing only the server addresses and closures for decreasing the active connection count
var onlineMut sync.RWMutex
var activeConnMut sync.Mutex

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
	Online     bool   // for health checks
}
type ServerConn struct {
	ServerAddr string
	ConnEnd    *func()
	ConnBegin  *func()
	IsOnline   *bool
}

func (s *server) connEnd() *func(method int) {
	var connEnd = func(method int) {
		if method == RR {
			activeConnMut.Lock()
			s.ActiveConn--
			activeConnMut.Unlock()
		}
	}
	return &connEnd
}

func (s *server) connBegin() *func() {
	var connBegin = func() {
		activeConnMut.Lock()
		s.ActiveConn++
		activeConnMut.Unlock()
	}
	return &connBegin
}

func Init(method int) chan ServerConn {
	serverStream = make(chan ServerConn, 10000) //buffer size

	jsonFile, err := os.Open("./servers.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	byteVal, _ := io.ReadAll(jsonFile)
	err2 := json.Unmarshal(byteVal, &servers)
	if err2 != nil {
		return nil
	}
	err1 := jsonFile.Close()
	if err1 != nil {
		fmt.Println(err1)
	}

	go healthCheckInit(servers, &onlineMut)

	time.Sleep(5 * time.Second) // use a better approach, maybe some type of signalling that initial checks are complete

	getNextServer(method)

	return serverStream
}

func getNextServer(method int) {
	switch method {
	case RR:
		go roundRobinInit(serverStream, servers, &onlineMut)
		break
	case LC:
		go leastConnInit(serverStream, servers, &onlineMut)
		break
	default:
		break
	}
}
