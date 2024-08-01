package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

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
	port          string
	templatesPath string
	paths         map[string][]Path
}

type Path struct {
	url     string
	method string
	value  string
}

type Payload struct {
	code    int
	content string
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

func (server *Server) listen() {
	// panic("Not implemented yet")
  _, err := net.Listen("tcp", server.port)
  if err != nil {
    panic("Couldn't start tcp server.")
  }
	return
}

func createServer(port string, templatesPath string, paths []Path) (Server, error) {
	var server Server
	server.port = port
	server.templatesPath = templatesPath
	for i := range len(paths) {
		j := 0 // HACK: iterate from 0 to 4
		if server.paths[paths[i].url][j].method == paths[i].method {
			errorMessage := fmt.Sprintf("Trying to override existing url %s for method %s with value %s, make sure your config is correct/", paths[i].url, paths[i].method, paths[i].value)
			panic(errorMessage)
		}
		// TODO: check if url exists
		// if yes, append to method and value
		// else create new path
		if server.paths[paths[i].url] != nil {
			server.paths[paths[i].url] = append(server.paths[paths[i].url], createPath(paths[i].url, paths[i].method, paths[i].value))
		}
		server.paths[paths[i].url] = append(server.paths[paths[i].url], createPath(paths[i].url, paths[i].method, paths[i].value))
	}
	return server, nil
}

// Default server on port 1337 with templatse path /templates and get root path /
// TODO: update to new server architecture (paths having arrays)
func createDefaultServer() (Server, error) {
	var server Server
	server.port = "1337"
	server.templatesPath = "/templates"
	server.paths["/"] = append(server.paths["/"], createPath("/", "GET", "Hello, World!"))
	return server, nil
}

// Returns new path object
func createPath(url string, method string, value string) Path {
	var path Path
  path.url = url
	path.method = method
	path.value = value
	return path
}

func createPathManyMethods(urls []string, methods []string, values []string) []Path {
	var paths []Path
	for i := range len(methods) {
    paths[i].url = urls[i]
		paths[i].method = methods[i]
		paths[i].value = values[i]
	}
	return paths
}

// TODO: update to new server architecture (paths having arrays)
func (server *Server) addPath(url string, method string, returnValue string) error {
	if strings.HasSuffix(returnValue, ".html") {
		// return html
	}
	htmlFileContent, err := os.ReadFile(server.templatesPath + "/" + returnValue)
	if err != nil {
		panic("File doesn't exist or has incorrect access permissions.")
	}
	htmlFile := string(htmlFileContent)
	server.paths[url] = append(server.paths[url], createPath(url, method, htmlFile))
	panic("Not implemented yet")
	return nil
}

func (server *Server) GET(path string) (Payload, error) {
	panic("Not implemented yet")
  return Payload{code: HTTP_OK, content: "No content"}, nil
}

func (server *Server) POST(path string, payload string) (Payload, error) {
	panic("Not implemented yet")
  return Payload{code: HTTP_OK, content: "No content"}, nil
}

func (server *Server) changePermission(path string, rule string) error {
	panic("Not implemented yet")
	return nil
}
