[![progress-banner](https://backend.codecrafters.io/progress/http-server/6f27fabe-0738-464a-8665-9d7b62191a86)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

# Simple Go HTTP Server

This is a basic HTTP server written in Go. It listens on a specified port and handles various HTTP requests, including serving files, echoing content, and returning the User-Agent string.

## Features

- **Echo Endpoint:** Returns the content appended to `/echo/` in the URL.
- **User-Agent Endpoint:** Returns the User-Agent string from the request.
- **File Serving:** Serves files from a specified directory. Supports `GET` and `POST` methods.
- **Gzip Compression:** Supports gzip compression based on the `Accept-Encoding` header.

## Usage

### Starting the Server

To start the server, run the following command:

```sh
go run main.go [directory]
```

- **Directory (optional):** The directory from which to serve files. If not specified, the current directory will be used.

## Endpoints
### Echo Endpoint

- **URL:** `/echo/{content}`
- **Method:** `GET`
- **Description:** Returns the `{content}` appended to the URL.

### User-Agent Endpoint

- **URL:** `/user-agent`
- **Method:** `GET`
- **Description:** Returns the User-Agent string from the request.

### File Serving

- **URL:** `/files/{filename}`
- **Method:** 
   - **GET:** Serves the specified file.
   - **POST:** Saves the content in the request body to the specified file.
- **Description:** Serves or saves files in the specified directory.

## Example Requests
### Echo Content:
```sh
curl http://localhost:4221/echo/hello
```

### Response:
```sh
hello
```
### Get User-Agent:
```sh
curl http://localhost:4221/user-agent
```

### Response:
```sh
curl/7.68.0
```
### Serve File:
```sh
curl http://localhost:4221/files/example.txt
```

### Save File:
```sh
curl -X POST -H "Content-Type: text/plain" -d "This is a test" http://localhost:4221/files/example.txt
```

## Implementation
Here's a brief overview of the code structure:
- **main.go:** The main entry point of the server.
   - **main():** Starts the server and listens on the specified port.
   - **handleConnection(conn net.Conn):** Handles incoming connections.
   - **parseRequest(conn net.Conn):** Parses the HTTP request.
   - **getStatus(statusCode int, statusText string):** Returns the HTTP status line.
   - **getDirectoryFromArgs():** Returns the directory from which to serve files, defaulting to the current directory if not specified.

## Error Handling
- Returns 400 Bad Request for malformed requests.
- Returns 404 Not Found for unknown paths or missing files.
- Returns 405 Method Not Allowed for unsupported methods.
- Returns 500 Internal Server Error for server-side errors during file operations.

## Gzip Compression
The server supports gzip compression if the `Accept-Encoding` header includes `gzip`. Compressed responses include the `Content-Encoding: gzip` header.

## License
This project is licensed under the MIT License.

