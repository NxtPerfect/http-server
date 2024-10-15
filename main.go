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
}

type Path struct {
	url    string
	method string
	value  string
}

type Response struct {
	code   int
	string string
}

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
		checkIsEof(err)
		fmt.Println("Error reading response:", err)
		return
	}

	getString, postString, httpString, hostString := server.getRequestValidationStrings()
	reqString := string(buff[:])

	if server.debug {
		fmt.Println("Got request:", reqString)
	}

	responseCode, responseString := getBadResponseValues()

	if reqStringIsValidGetOrPost(reqString, getString, postString, httpString, hostString) {
		responseCode = HTTP_OK
		responseString = "OK"
	}

	relativeFilePath, err := getRelativeFilePathOfRequest(reqString, conn)
	if err != nil {
		return
	}

	filePathRune := []rune(relativeFilePath)
	httpIndex, err := getIndexOfHttpInRequestOrSendReplyToConnection(relativeFilePath, conn)
	if err != nil {
		return
	}

	filePath := string(filePathRune[:httpIndex])
	if !server.isValidPath(filePath) {
		writeBadResponseToConnection(filePath, conn)
		return
	}

	relativeFilePath = server.createRelativeFilePath(filePath)

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

	response := Response{code: responseCode, string: responseString}
	writeGoodResponseToConnection(response, contentLength, requestFile, conn)
}

func checkIsEof(err error) {
	if err.Error() == "EOF" {
		fmt.Println("Closing connection, got no response to read. Error:", err)
		return
	}
}

func (server *Server) getRequestValidationStrings() (string, string, string, string) {
	getString := fmt.Sprintf("GET")
	postString := fmt.Sprintf("POST")
	httpString := fmt.Sprintf("HTTP/1.")
	hostString := fmt.Sprintf("Host: %s:%s", server.host, server.port)
	return getString, postString, httpString, hostString
}

func getBadResponseValues() (int, string) {
	return HTTP_BAD_REQUEST, "BAD REQUEST"
}

func reqStringIsValidGetOrPost(reqString string, getString string, postString string, httpString string, hostString string) bool {
	return (strings.Contains(reqString, getString) || strings.Contains(reqString, postString)) &&
		strings.Contains(reqString, httpString) && strings.Contains(reqString, hostString)
}

func getRelativeFilePathOfRequest(reqString string, conn net.Conn) (string, error) {
	_, relativeFilePath, found := strings.Cut(reqString, "GET ")
	if found == false {
		_, relativeFilePath, found = strings.Cut(reqString, "POST ")
		if found == false {
			response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", HTTP_BAD_REQUEST, "BAD REQUEST")
			conn.Write([]byte(response))
			fmt.Println("Path not found in get")
			return "", fmt.Errorf("Path not found in get or post. Check if request is valid.")
		}
	}
	return relativeFilePath, nil
}

func getIndexOfHttpInRequestOrSendReplyToConnection(relativeFilePath string, conn net.Conn) (int, error) {
	httpIndex := strings.Index(relativeFilePath, " HTTP/")
	if httpIndex == -1 {
		response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", HTTP_BAD_REQUEST, "BAD REQUEST")
		conn.Write([]byte(response))
		fmt.Println("Bad request")
		return -1, fmt.Errorf("Request doesn't have HTTP/ string. Check if request is valid.")
	}
	return httpIndex, nil
}

func writeBadResponseToConnection(filePath string, conn net.Conn) {
	responseCode := HTTP_NOT_FOUND
	responseString := "NOT FOUND"
	response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", responseCode, responseString)
	response += fmt.Sprintf("Server: Custom/Server\r\n")

	conn.Write([]byte(response))

	fmt.Printf("Path not in server paths %s.\n", filePath)
	return
}

func (server *Server) createRelativeFilePath(filePath string) string {
	return "." + server.templatesPath + "/" + server.getFileFromPath(filePath)
}

func writeGoodResponseToConnection(response Response, contentLength int64, requestFile []byte, conn net.Conn) {
	connResponse := fmt.Sprintf("HTTP/1.1 %v %s\r\n", response.code, response.string)
	connResponse += fmt.Sprintf("Server: Custom/Server\r\n")
	connResponse += fmt.Sprintf("Content-Type: text/html\r\n")
	connResponse += fmt.Sprintf("Content-Length: %d\r\n\n", contentLength)
	connResponse += fmt.Sprintf(string(requestFile))

	conn.Write([]byte(connResponse))
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
