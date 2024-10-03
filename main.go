package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	GET = iota
	POST
	PUT
	DELETE
)

type Server struct {
	host          string
	port          string
	templatesPath string
	paths         map[string][]Path
	readyChan     chan struct{}
	shutdownChan  chan struct{}
	debug         bool
	// wg            *sync.WaitGroup
}

type Path struct {
	url    string
	method string
	value  string
}

// General HTTP Response type
// code is HTTP_ code
// content is the actual http response
type Payload struct {
	code    int
	content string
}

// HTTP response codes as int values
const (
	HTTP_OK                    = 200
	HTTP_ACCEPTED              = 202
	HTTP_BAD_REQUEST           = 400
	HTTP_UNAUTHORIZED          = 401
	HTTP_FORBIDDEN             = 403
	HTTP_NOT_FOUND             = 404
	HTTP_GONE                  = 410
	HTTP_INTERNAL_SERVER_ERROR = 500
)

func (server *Server) Listen() error {
	server.host = "127.0.0.1"
	ln, err := net.Listen("tcp", server.host+":"+server.port)

	defer func() {
		if ln != nil {
			if server.debug {
				fmt.Println("Closing down server")
			}
			ln.Close()
		}
	}()

	if err != nil {
		fmt.Printf("Couldn't listen to port %s %s", server.port, err)
		return err
	}

	close(server.readyChan)

	fmt.Printf("Accepting connections on %s:%s\n", server.host, server.port)

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				break
			}
			go server.handleConnection(conn)
		}
	}()

	<-server.shutdownChan
	return nil
}

func (server *Server) handleConnection(conn net.Conn) {
	defer func() {
		if server.debug {
			fmt.Println("Closing the connection server-side")
		}
		conn.Close()
	}()
	if server.debug {
		fmt.Println("New connection.")
	}

	buff := make([]byte, 32768)
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, err := conn.Read(buff)

	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("Closing connection, got no response to read. Error:", err)
			return
		}
		fmt.Println("Error reading response:", err)
		panic("Couldn't read response.")
	}

	getString := fmt.Sprintf("GET")
	postString := fmt.Sprintf("POST")
	httpString := fmt.Sprintf("HTTP/1.")
	hostString := fmt.Sprintf("Host: %s:%s", server.host, server.port)
	reqString := string(buff[:])

	if server.debug {
		fmt.Println("Got request:", reqString)
	}

	responseCode := HTTP_BAD_REQUEST
	responseString := "BAD REQUEST"

	if (strings.Contains(reqString, getString) || strings.Contains(reqString, postString)) &&
		strings.Contains(reqString, hostString) && strings.Contains(reqString, httpString) {
		responseCode = HTTP_OK
		responseString = "OK"
	}

	_, relativeFilePath, found := strings.Cut(reqString, "GET ")
	if found == false {
		_, relativeFilePath, found = strings.Cut(reqString, "POST ")
		if found == false {
			response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", HTTP_BAD_REQUEST, "BAD REQUEST")
			conn.Write([]byte(response))
			fmt.Println("Path not found in get")
			return
		}
	}

	filePathRune := []rune(relativeFilePath)
	httpIndex := strings.Index(relativeFilePath, " HTTP/")
	if httpIndex == -1 {
		response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", HTTP_BAD_REQUEST, "BAD REQUEST")
		conn.Write([]byte(response))
		fmt.Println("Bad request")
		return
	}

	if !server.isValidPath(string(filePathRune[:httpIndex])) {
		responseCode = HTTP_NOT_FOUND
		responseString = "NOT FOUND"
		response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", responseCode, responseString)
		response += fmt.Sprintf("Server: Custom/Server\r\n")

		conn.Write([]byte(response))

		fmt.Printf("Path not in server paths %s.\n", string(filePathRune[:httpIndex]))
		return
	}

	relativeFilePath = "." + server.templatesPath + "/" + server.getFileFromPath(string(filePathRune[:httpIndex]))

	fileinfo, err := os.Stat(relativeFilePath)
	if err != nil {
		fmt.Println("Couldn't read stats of file", relativeFilePath)
		return
	}

	contentLength := fileinfo.Size()
	requestFile, err := os.ReadFile(relativeFilePath)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return
	}

	response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", responseCode, responseString)
	response += fmt.Sprintf("Server: Custom/Server\r\n")
	response += fmt.Sprintf("Content-Type: text/html\r\n")
	response += fmt.Sprintf("Content-Length: %d\r\n\n", contentLength)
	response += fmt.Sprintf(string(requestFile))

	conn.Write([]byte(response))
}

func (server *Server) isValidPath(path string) bool {
	if server.paths[path] != nil {
		return true
	}
	fmt.Println(server.paths)
	return false
}

func (server *Server) getFileFromPath(path string) string {
	if server.paths[path] != nil {
		return server.paths[path][0].value
	}
	fmt.Println("No value from paths, returning index.html")
	return "index.html"
}

func CreateServer(host string, port string, templatesPath string, paths []Path, debug bool) (Server, func()) {
	var server Server
	server.host = host
	server.port = port
	server.templatesPath = templatesPath
	server.paths = make(map[string][]Path)
	for _, path := range paths {
		server.paths[path.url] = append(server.paths[path.url], Path{path.url, path.method, path.value})
	}
	server.readyChan = make(chan struct{})
	server.shutdownChan = make(chan struct{})
	server.debug = debug

	return server, func() {
		server.Shutdown()
		time.Sleep(100 * time.Millisecond)
	}
}

func CreateDefaultServer() (Server, func()) {
	var server Server
	server.host = "127.0.0.1"
	server.port = "1337"
	server.templatesPath = "/templates"
	var path Path = Path{url: "/", method: "GET", value: "index.html"}
	server.paths = make(map[string][]Path)
	server.paths["/"] = append(server.paths["/"], path)
	server.readyChan = make(chan struct{})
	server.shutdownChan = make(chan struct{})
	server.debug = true
	return server, func() {
		server.Shutdown()
		time.Sleep(100 * time.Millisecond)
	}
}

func (server *Server) Shutdown() {
	close(server.shutdownChan)
}

func (server *Server) AddPath(url string, method string, returnValue string) error {
	if strings.HasSuffix(returnValue, ".html") {
		// TODO: return html
	}
	htmlFileContent, err := os.ReadFile(server.templatesPath + "/" + returnValue)
	if err != nil {
		panic("File doesn't exist or has incorrect access permissions.")
	}
	htmlFile := string(htmlFileContent)
	server.paths[url] = append(server.paths[url], Path{url, method, htmlFile})

	return nil
}
