package balancer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

const RR = 1
const LC = 2

var servers []server // contains all server addresses from the JSON and their current active connection/s
var onlineMut sync.RWMutex
var activeConnMut sync.Mutex
var healthCheckWG sync.WaitGroup

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
	Online     bool   // for health checks
}
type ServerConn struct {
	ServerAddr string
	ConnEnd    *func()
	ConnBegin  *func(method int)
	IsOnline   *bool
}

func (s *server) connEnd() *func() {
	var connEnd = func() {
		activeConnMut.Lock()
		s.ActiveConn--
		activeConnMut.Unlock()

	}
	return &connEnd
}

func (s *server) connBegin() *func(method int) {
	var connBegin = func(method int) {
		if method == RR {
			activeConnMut.Lock()
			s.ActiveConn++
			activeConnMut.Unlock()
		}
	}
	return &connBegin
}

func Init(method int) chan ServerConn {
	healthCheckWG.Add(1)
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

	go healthCheckInit(servers, &activeConnMut, &onlineMut, &healthCheckWG)

	healthCheckWG.Wait()

	return getNextServer(method)
}

func getNextServer(method int) chan ServerConn {
	switch method {
	case RR:
		return roundRobinInit(servers, &onlineMut)
	case LC:
		return leastConnInit(servers, &activeConnMut, &onlineMut)
	default:
		return nil
	}
}
