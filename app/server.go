package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Handling new connection...")

	req, err := parseRequest(conn)
	if err != nil {
		fmt.Fprintf(conn, "HTTP/1.1 400 Bad Request\r\n\r\n%s", err.Error())
		fmt.Println("Error parsing request:", err)
		return
	}

	fmt.Printf("Received request: Method=%s, Path=%s, Headers=%v, UserAgent=%s, Body=%s\n", req.Method, req.Path, req.Headers, req.UserAgent, req.Body)

	var body string
	var responseHeaders string

	switch path := req.Path; {
	case strings.HasPrefix(path, "/echo/"):
		content := strings.TrimPrefix(path, "/echo/")
		body = content
		responseHeaders = fmt.Sprintf("%s\r\nContent-Type: text/plain", getStatus(200, "OK"))
		fmt.Printf("Echo response: %s\n", content)
	case path == "/user-agent":
		body = req.UserAgent
		responseHeaders = fmt.Sprintf("%s\r\nContent-Type: text/plain", getStatus(200, "OK"))
		fmt.Printf("User-Agent response: %s\n", req.UserAgent)
	case strings.HasPrefix(path, "/files/"):
		dir := getDirectoryFromArgs()
		fileName := strings.TrimPrefix(path, "/files/")
		filePath := filepath.Join(dir, fileName)
		fmt.Println("req.Method: ", req.Method)

		if req.Method == "GET" {
			file, err := os.ReadFile(filePath)
			if err != nil {
				responseHeaders = getStatus(404, "Not Found") + "\r\n\r\n"
				fmt.Printf("File not found: %s\n", filePath)
			} else {
				body = string(file)
				responseHeaders = fmt.Sprintf("%s\r\nContent-Type: application/octet-stream", getStatus(200, "OK"))
				fmt.Printf("File served: %s\n", filePath)
			}
		} else if req.Method == "POST" {
			contentLength, err := strconv.Atoi(req.Headers["Content-Length"])
			if err != nil {
				responseHeaders = getStatus(400, "Bad Request") + "\r\n\r\n"
				break
			}
			body := []byte(req.Body)
			if len(body) != contentLength {
				responseHeaders = getStatus(400, "Bad Request") + "\r\n\r\n"
				break
			}
			err = os.WriteFile(filePath, body, 0644)
			if err != nil {
				responseHeaders = getStatus(500, "Internal Server Error") + "\r\n\r\n"
			} else {
				responseHeaders = getStatus(201, "Created") + "\r\n\r\n"
				fmt.Printf("File created: %s\n", filePath)
			}
		} else {
			responseHeaders = getStatus(405, "Method Not Allowed") + "\r\n\r\n"
		}
	case path == "/":
		responseHeaders = getStatus(200, "OK") + "\r\n\r\n"
		fmt.Println("Root path response: 200 OK")
	default:
		responseHeaders = getStatus(404, "Not Found") + "\r\n\r\n"
		fmt.Printf("Path not found: %s\n", path)
	}

	// Handle gzip compression based on Accept-Encoding header
	acceptEncoding := req.Headers["Accept-Encoding"]
	if strings.Contains(acceptEncoding, "gzip") {
		var compressedBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressedBody)
		_, err := gzipWriter.Write([]byte(body))
		if err != nil {
			fmt.Println("Error compressing body:", err)
			compressedBody.Reset()
			compressedBody.WriteString(body)
		}
		gzipWriter.Close()
		responseHeaders += fmt.Sprintf("\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n", compressedBody.Len())
		conn.Write([]byte(responseHeaders))
		conn.Write(compressedBody.Bytes())
	} else {
		responseHeaders += fmt.Sprintf("\r\nContent-Length: %d\r\n\r\n", len(body))
		conn.Write([]byte(responseHeaders))
		conn.Write([]byte(body))
	}
	fmt.Println("Response sent to client")
}

func parseRequest(conn net.Conn) (*HTTPRequest, error) {
	reader := bufio.NewReader(conn)
	var req HTTPRequest
	req.Headers = make(map[string]string)
	lineNum := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading request: %v", err)
		}
		line = strings.Trim(line, "\r\n")
		fmt.Printf("Parsing line: %s\n", line)

		if lineNum == 0 {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid request line")
			}
			req.Method = parts[0]
			req.Path = parts[1]
		} else {
			if line == "" {
				break
			}
			headers := strings.SplitN(line, ": ", 2)
			if len(headers) < 2 {
				continue
			}
			if headers[0] == "User-Agent" {
				req.UserAgent = headers[1]
			}
			req.Headers[headers[0]] = headers[1]
		}
		lineNum++
	}

	if req.Method == "POST" {
		contentLength, err := strconv.Atoi(req.Headers["Content-Length"])
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length")
		}
		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %v", err)
		}
		req.Body = string(body)
	}

	return &req, nil
}

func getStatus(statusCode int, statusText string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s", statusCode, statusText)
}

func getDirectoryFromArgs() string {
	if len(os.Args) > 2 {
		return os.Args[2]
	}
	return "." // Default directory
}