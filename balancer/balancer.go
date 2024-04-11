package balancer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const RR = 1
const LC = 2
const HASH = 3

type server struct {
	Address    string `json:"address"`
	ActiveConn uint   `json:"active_conn,omitempty"`
}

var servers []server
var rrPtr = 0

func Init() {
	jsonFile, err := os.Open("../servers.json")
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

// Servers[i].ActiveConn
// Servers[i].Address

func GetNextServer(method int) string {
	switch method {
	case RR:
		return rrNext()
	case LC:
		return lcNext()
	case HASH:
		return hashNext()
	}
}

func rrNext() string {
	nextServer := rrPtr
	rrPtr++
	return servers[nextServer].Address
}

func lcNext() string {

}

func hashNext() string {

}
