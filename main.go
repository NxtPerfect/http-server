package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

/*
  HTTP 1.0 Response
  HTTP/1.0 200 OK
  Date: Fri, 08 Aug 2003 08:12:31 GMT
  Server: Apache/1.3.27 (Unix)
  MIME-version: 1.0
  Last-Modified: Fri, 01 Aug 2003 12:45:26 GMT
  Content-Type: text/html
  Content-Length: 2345
  ** a blank line *
  <HTML> ...
*/

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
	shutdownChan  chan struct{}
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
			fmt.Println("Closing down server")
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
		fmt.Println("Closing the connection server-side")
		conn.Close()
	}()
	fmt.Println("New connection.")

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

	getString := fmt.Sprintf("GET / HTTP/1.")
	postString := fmt.Sprintf("POST / HTTP/1.")
	hostString := fmt.Sprintf("Host: %s:%s", server.host, server.port)
	reqString := string(buff[:])

	fmt.Println("Got request:", reqString)

	responseCode := 0
	responseString := ""

	if (strings.Contains(reqString, getString) || strings.Contains(reqString, postString)) &&
		strings.Contains(reqString, hostString) {
		responseCode = HTTP_OK
		responseString = "OK"
	} else {
		responseCode = HTTP_BAD_REQUEST
		responseString = "BAD REQUEST"
	}
	fmt.Println("Bad request!")
	conn.Write([]byte(fmt.Sprintf("%d %s", responseCode, responseString)))
	// responseCode = HTTP_BAD_REQUEST
	// responseString = "BAD REQUEST"
	// date := time.Now().UTC().Format("02 Jan 2006 15:04:05 MST")
	//
	// _, relativeFilePath, found := strings.Cut(reqString, "GET ")
	// if found == false {
	// 	fmt.Println("Path not found in get")
	// 	return
	// }
	// filePathRune := []rune(relativeFilePath)
	// httpIndex := strings.Index(relativeFilePath, " HTTP/")
	// if httpIndex == -1 {
	// 	fmt.Println("Bad request")
	// 	return
	// }
	// relativeFilePath = string(filePathRune[:httpIndex])
	//
	// fileinfo, err := os.Stat(relativeFilePath)
	// if err != nil {
	// 	return
	// }
	// mtime := fileinfo.Sys().(*syscall.Stat_t).Mtim
	//
	// fmt.Println("Modified time:", mtime)
	//
	// // lastModified := time.Format("02 Jan 2006 15:04:05 MST")
	//
	// contentLength := fileinfo.Size()
	// requestFile, err := os.ReadFile(server.templatesPath + relativeFilePath)
	// if err != nil {
	// 	fmt.Println("Failed to read file:", err)
	// 	return
	// }
	//
	// response := fmt.Sprintf("HTTP/1.1 %v %s\r\n", responseCode, responseString)
	// response += fmt.Sprintf("Date: %s\r\n", date)
	// response += fmt.Sprintf("Server: Custom/Server\r\n")
	// // response += fmt.Sprintf("Last-Modified: %s\r\n", lastModified)
	// response += fmt.Sprintf("Content-Type: text/html\r\n")
	// response += fmt.Sprintf("Content-Length: %d\r\n\n", contentLength)
	// response += fmt.Sprintf(string(requestFile))
	//
	// conn.Write([]byte(response))
}

func CreateServer(host string, port string, templatesPath string, paths []Path) (Server, func()) {
	var server Server
	server.host = host
	server.port = port
	server.templatesPath = templatesPath
	server.paths = make(map[string][]Path)
	for i := range len(paths) {
		server.paths[paths[i].url] = append(server.paths[paths[i].url], Path{paths[i].url, paths[i].method, paths[i].value})
	}
	server.readyChan = make(chan struct{})
	server.shutdownChan = make(chan struct{})
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
	var path Path = Path{url: "/", method: "GET", value: "Hello, World!"}
	server.paths = make(map[string][]Path)
	server.paths["/"] = append(server.paths["/"], path)
	server.readyChan = make(chan struct{})
	server.shutdownChan = make(chan struct{})
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
