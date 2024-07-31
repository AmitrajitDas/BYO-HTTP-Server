package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type HTTPRequest struct {
	Method    string
	Path      string
	Headers   map[string]string
	Body      string
	UserAgent string
}

func main() {
	fmt.Println("Starting the server...")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Server is listening on port 4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		fmt.Println("Connection established!")
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling new connection...")

	scanner := bufio.NewScanner(conn)
	req, _ := parseStatus(scanner)
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(conn, "reading standard input:", err)
		fmt.Println("Error reading from connection:", err)
		return
	}

	fmt.Printf("Received request: Method=%s, Path=%s, Headers=%s, UserAgent=%s\n", req.Method, req.Path, req.Headers, req.UserAgent)

	var response string
	switch path := req.Path; {
	case strings.HasPrefix(path, "/echo/"):
		content := strings.TrimLeft(path, "/echo/")
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(content), content)
		fmt.Printf("Echo response: %s\n", content)
	case path == "/user-agent":
		response = fmt.Sprintf("%s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", getStatus(200, "OK"), len(req.UserAgent), req.UserAgent)
		fmt.Printf("User-Agent response: %s\n", req.UserAgent)
	case path == "/":
		response = getStatus(200, "OK") + "\r\n\r\n"
		fmt.Println("Root path response: 200 OK")
	default:
		response = getStatus(404, "Not Found") + "\r\n\r\n"
		fmt.Printf("Path not found: %s\n", path)
	}

	conn.Write([]byte(response))
	fmt.Println("Response sent to client")
}

func parseStatus(scanner *bufio.Scanner) (*HTTPRequest, error) {
	var req HTTPRequest
	req.Headers = make(map[string]string)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text()
		fmt.Printf("Parsing line: %s\n", line)
		if i == 0 {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid request line")
			}
			req.Method = parts[0]
			req.Path = parts[1]
			continue
		}
		headers := strings.Split(line, ": ")
		if len(headers) < 2 {
			req.Body = line
			break
		}
		if headers[0] == "User-Agent" {
			req.UserAgent = headers[1]
		}
		req.Headers[headers[0]] = headers[1]
	}
	return &req, nil
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}
