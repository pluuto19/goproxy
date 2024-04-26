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
var mut sync.RWMutex

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
	Online     bool   // for health checks
}
type ServerConn struct {
	ServerAddr string
	ConnEnd    *func()
	ConnBegin  *func()
}

func (s *server) connEnd() *func() {
	var connEnd = func() {
		fmt.Println("about to decrement from " + s.Address)
		fmt.Println(time.Now().Format(time.RFC3339Nano))
		s.ActiveConn--
	}
	return &connEnd
}

func (s *server) connBegin() *func() {
	var connBegin = func() {
		fmt.Println("about to increase " + s.Address)
		fmt.Println(time.Now().Format(time.RFC3339Nano))
		s.ActiveConn++
	}
	return &connBegin
}

func Init(method int) chan ServerConn {
	serverStream = make(chan ServerConn, 10000)

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
	go healthCheckInit(servers, &mut)
	getNextServer(method)

	return serverStream
}

func getNextServer(method int) {
	switch method {
	case RR:
		go roundRobinInit(serverStream, servers, &mut)
		break
	case LC:
		go leastConnInit(serverStream, servers, &mut)
		break
	default:
		break
	}
}
