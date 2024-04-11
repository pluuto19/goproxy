package balancer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

const RR = 1
const LC = 2

//const HASH = 3

var servers []server
var rrPtr = 0

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
}

func (s *server) connEnd() func() {
	return func() {
		fmt.Println("about to decrement from " + s.Address)
		s.ActiveConn--
	}
}

func Init() {
	jsonFile, err := os.Open("./servers.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	byteVal, _ := io.ReadAll(jsonFile)
	err2 := json.Unmarshal(byteVal, &servers)
	if err2 != nil {
		return
	}

	err1 := jsonFile.Close()
	if err1 != nil {
		fmt.Println(err1)
	}
}

func GetNextServer(method int) (serverAddr string, ConnEnd func()) {
	switch method {
	case RR:
		return rrNext()
	case LC:
		return lcNext()
	//case HASH:
	//	return hashNext()
	default:
		return "", nil
	}
}

func rrNext() (string, func()) {
	thisServer := rrPtr
	rrPtr++

	if rrPtr%len(servers) == 0 {
		rrPtr = 0
	}

	servers[thisServer].ActiveConn++
	return servers[thisServer].Address, servers[thisServer].connEnd()
}

func lcNext() (string, func()) {
	for _, elem := range servers {
		fmt.Println(elem.Address, " ", elem.ActiveConn)
	}
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].ActiveConn < servers[j].ActiveConn
	})
	for _, elem := range servers {
		fmt.Println(elem.Address, " ", elem.ActiveConn)
	}
	serverAddr := servers[0].Address
	connEnd := servers[0].connEnd()
	servers[0].ActiveConn++
	for _, elem := range servers {
		fmt.Println(elem.Address, " ", elem.ActiveConn)
	}
	return serverAddr, connEnd
}

//func hashNext() (string, func()) {
//	return
//}
