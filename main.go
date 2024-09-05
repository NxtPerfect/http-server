package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Http response header
type Head struct {
	ContentType     mediaType
	ContentEncoding contentCoding
	ContentLanguage string
	ContentLocation string
}

type mediaType struct {
	// Case insensitive
	Type      string // Type "/" Subtype *(OWS ";" OWS parameter)
	Subtype   string
	parameter string // token "=" ( token / quoted-string )
	/*
	   First is best, but all are accepted
	    text/html;charset=utf-8
	    text/html;charset=UTF-8
	    Text/HTML;Charset="utf-8"
	    text/html; charset="utf-8"
	*/
}

type charset struct {
	token string
}

type contentCoding struct {
	token string // "compress" "x-compress" "deflate" "gzip" "x-gzip"
}

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

// Used to connect to port and listen for connections
func (server *Server) Listen() {
	server.host = "127.0.0.1"
	ln, err := net.Listen("tcp", server.host+":"+server.port)
	// Causes nil pointer issue, but without that
	// we never free the ip+port
	defer ln.Close()

	if err != nil {
		fmt.Printf("Couldn't listen to port %s %s", server.port, err)
		panic(err)
	}

	close(server.readyChan)

	for {
		fmt.Printf("Accepting connections on %s:%s\n", server.host, server.port)
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Got new connection. Closing...\n")

	buff := make([]byte, 32768)
	_, err := conn.Read(buff)
	if err != nil {
		if err.Error() == "EOF" {
			fmt.Println("Closing connection, got no response to read.")
			return
		}
		panic("Couldn't read response")
		fmt.Println("Got an error reading response:", err)
		return
	}

	hostString := fmt.Sprintf("Host: %s:%s", server.host, server.port)
	getString := fmt.Sprintf("GET / HTTP/1.1")
	postString := fmt.Sprintf("POST / HTTP/1.1")
	reqString := string(buff[:])

	if (!strings.Contains(reqString, getString) && !strings.Contains(reqString, postString)) ||
		!strings.Contains(reqString, hostString) {
		fmt.Println("Bad request!")
		conn.Write([]byte(fmt.Sprintf("%d", HTTP_BAD_REQUEST)))
		return
	}
	conn.Write([]byte(fmt.Sprintf("%d", HTTP_OK)))
}

func CreateServer(port string, templatesPath string, paths []Path) (Server, error) {
	var server Server
	// server.host = host
	server.port = port
	server.templatesPath = templatesPath
	server.paths = make(map[string][]Path)
	for i := range len(paths) {
		server.paths[paths[i].url] = append(server.paths[paths[i].url], Path{paths[i].url, paths[i].method, paths[i].value})
	}
	return server, nil
}

// Default server on port 80 with templatse path /templates and get root path /
// TODO: update to new server architecture (paths having arrays)
func CreateDefaultServer() (Server, error) {
	var server Server
	// server.host = "127.0.0.1"
	server.port = "1337"
	server.templatesPath = "/templates"
	var path Path = Path{url: "/", method: "GET", value: "Hello, World!"}
	// server.paths["/"] = append(server.paths["/"], createPath("/", "GET", "Hello, World!"))
	server.paths = make(map[string][]Path)
	server.paths["/"] = append(server.paths["/"], path)
	return server, nil
}

// Adds new url path to server
// returns an error
// TODO: update to new server architecture (paths having arrays)
func (server *Server) AddPath(url string, method string, returnValue string) error {
	if strings.HasSuffix(returnValue, ".html") {
		// return html
	}
	htmlFileContent, err := os.ReadFile(server.templatesPath + "/" + returnValue)
	if err != nil {
		panic("File doesn't exist or has incorrect access permissions.")
	}
	htmlFile := string(htmlFileContent)
	server.paths[url] = append(server.paths[url], Path{url, method, htmlFile})
	panic("Not implemented yet %s")
	println(htmlFile)
	return nil
}

// Send payload to url
// returns reponse and error
func (server *Server) GET(path string) (Payload, error) {
	// panic("Get request not implemented yet")
	//
	// rt := fmt.Sprintf("GET %v HTTP/1.1\r\n", path)
	// rt += fmt.Sprintf("Host: %v\r\n", server.host)
	// rt += fmt.Sprintf("Connection: close\r\n")
	// rt += fmt.Sprintf("\r\n")
	//
	// _, err = conn.Write([]byte(rt))
	// if err != nil {
	// 	fmt.Print(err)
	// }
	return Payload{code: HTTP_OK, content: "No content"}, nil
}

// Send payload to url
// returns response and error
func (server *Server) POST(url string, payload string) (Payload, error) {
	// panic("Post request not implemented yet")
	return Payload{code: HTTP_OK, content: "No content"}, nil
}

// Set permission for path
// checks if user sends correct header with role
func (server *Server) ChangePermission(path string, rule string) error {
	// panic("Changing permissions not implemented yet")
	return nil
}
