package balancer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

const RR = 1
const LC = 2

var servers []server             // contains all server addresses from the JSON and their current active connection/s
var rrPtr = 0                    // for selecting the next server in Round Robin fashion
var serverStream chan ServerConn // for sharing only the server addresses and closures for decreasing the active connection count

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
}
type ServerConn struct {
	ServerAddr string
	ConnEnd    *func()
}

func (s *server) connEnd() *func() {
	var connEnd = func() {
		fmt.Println("about to decrement from " + s.Address)
		fmt.Println(time.Now().Format(time.RFC3339Nano))
		s.ActiveConn--
	}
	return &connEnd
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
	getNextServer(method)

	return serverStream
}

func getNextServer(method int) {
	switch method {
	case RR:
		go rrNext()
		break
	case LC:
		go lcNext()
		break
	default:
		break
	}
}

func rrNext() {
	defer close(serverStream)
	for {
		servers[rrPtr].ActiveConn++
		serverStream <- ServerConn{servers[rrPtr].Address, servers[rrPtr].connEnd()}
		rrPtr++
		if rrPtr%len(servers) == 0 {
			rrPtr = 0
		}
	}
}

func lcNext() {
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].ActiveConn < servers[j].ActiveConn
	})
	//serverAddr := servers[0].Address
	//connEnd := servers[0].connEnd()
	servers[0].ActiveConn++
	//return serverAddr, connEnd
}

// min heap for LC
// server health
// using cond to sleep the xxNext to save on OS thread scheduling when the buffered channel becomes full
