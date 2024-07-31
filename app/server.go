package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connection established!")

	req := make([]byte, 1024)
	conn.Read(req)

	splitHeader := strings.Split(string(req), "\r\n")
	fmt.Println("splitHeader: ", splitHeader)
	splitRequestLine := strings.Split(splitHeader[0], " ")
	fmt.Println("splitRequestLine: ", splitRequestLine)

	if splitRequestLine[1] == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.Split(splitRequestLine[1], "/")[1] == "echo" {
		reqBody := strings.Split(splitRequestLine[1], "/")[2]
		fmt.Println("reqBody: ", reqBody)
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(len(reqBody)) + "\r\n\r\n" + reqBody))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
