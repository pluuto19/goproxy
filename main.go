package main

import (
	"fmt"
	"net"
)

const bufSize = 1536

var backendServers = []string{"localhost:8008"}

func main() {
	serverSpec, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	if err != nil {
		return
	}
	welcSock, err := net.ListenTCP("tcp", serverSpec) //welcoming socket takes in a
	if err != nil {                                   //server specification struct
		return
	}
	fmt.Println("Proxy running ...")
	serveConcurrRequest(welcSock.Accept())
}
func serveConcurrRequest(clientConnSock net.Conn, err error) {

	if err != nil {
		return
	}

	clientRecvBuffer := make([]byte, bufSize)

	n, err := clientConnSock.Read(clientRecvBuffer)
	if err != nil {
		return
	}

	//---------- Call a backend server and send it HTTP request from client ---------//

	tcpAddr, err := net.ResolveTCPAddr("tcp4", backendServers[0])
	if err != nil {
		return
	}
	backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return
	}
	_, err1 := backendConnSock.Write(clientRecvBuffer[0:n])
	if err1 != nil {
		return
	}
	backendRecvBuffer := make([]byte, bufSize)
	m, err := backendConnSock.Read(backendRecvBuffer)
	if err != nil {
		return
	}
	err2 := backendConnSock.Close()
	if err2 != nil {
		return
	}
	_, err3 := clientConnSock.Write(backendRecvBuffer[0:m])
	if err3 != nil {
		return
	}
	err4 := clientConnSock.Close()
	if err4 != nil {
		return
	}
}

// loop or no loop because of Transport Layer segmentation
