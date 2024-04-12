package main

import (
	"fmt"
	"loadbalancerproxy/balancer"
	"net"
)

const bufSize = 1536

func main() {
	serverStream := balancer.Init(balancer.RR)

	serverSpec, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	if err != nil {
		return
	}
	welcSock, err := net.ListenTCP("tcp", serverSpec) //welcoming socket takes in a server specification struct
	if err != nil {
		return
	}
	fmt.Println("Proxy running ...")
	for {
		clientConnSock, err := welcSock.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go serveConcurrRequest(serverStream, clientConnSock)
	}
}
func serveConcurrRequest(serverStream chan balancer.ServerConn, clientConnSock net.Conn) {

	clientRecvBuffer := make([]byte, bufSize)

	n, err := clientConnSock.Read(clientRecvBuffer)
	if err != nil {
		fmt.Println(err)
		return
	}

	serverConn, ok := <-serverStream
	if !ok {
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", serverConn.ServerAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	backendConnSock, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err1 := backendConnSock.Write(clientRecvBuffer[0:n])
	if err1 != nil {
		fmt.Println(err1)
		return
	}
	backendRecvBuffer := make([]byte, bufSize)
	m, err := backendConnSock.Read(backendRecvBuffer)
	if err != nil {
		fmt.Println(err)
		return
	}
	err2 := backendConnSock.Close()

	(*serverConn.ConnEnd)()

	if err2 != nil {
		fmt.Println(err2)
		return
	}
	_, err3 := clientConnSock.Write(backendRecvBuffer[0:m])
	if err3 != nil {
		fmt.Println(err3)
		return
	}
	err4 := clientConnSock.Close()
	if err4 != nil {
		fmt.Println(err4)
		return
	}
}
