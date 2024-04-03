package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run client.go <port> <resource_name>")
		return
	}

	port := os.Args[1]
	resourceName := os.Args[2]

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		fmt.Println("Closing connection: ", conn.RemoteAddr())
		if err != nil {

		}
	}(conn)

	_, err = conn.Write([]byte(resourceName))
	if err != nil {
		fmt.Println("Error sending data:", err)
		return
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error receiving data:", err)
		return
	}

	resourceValue := string(buffer[:n])
	fmt.Println("Received resource value:\n", resourceValue)
}
