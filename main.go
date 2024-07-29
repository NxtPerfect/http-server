package main

import "testing"

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

type Server struct {
  port int
}

const (
  HTTP_OK = 200
  HTTP_ACCEPTED = 202
  HTTP_BAD_REQUEST = 400
  HTTP_UNAUTHORIZED = 401
  HTTP_FORBIDDEN = 403
  HTTP_NOT_FOUND = 404
  HTTP_GONE = 410
  HTTP_INTERNAL_SERVER_ERROR = 500
)

func testValidPath(t *testing.T) {
  server := createServer()
  server.run()
  code := server.GET("/")
  if code != HTTP_OK {
    t.Fatalf(`Requested GET for '/' failed with %d`, code)
  }
}

func testUnauthorized(t *testing.T) {
  server := createServer()
  server.addPath("/admin")
  server.changePermission("/admin", "Role: Admin")
  server.run()
  code := server.GET("/admin")
  if code != HTTP_UNAUTHORIZED {
    t.Fatalf(`Requested GET for '/admin' without permissions accepted with %d`, code)
  }
}

func testPostWithPayload(t *testing.T) {
  server := createServer()
  server.addPath("/login", "POST")
  server.run()
  code := server.POST("/login", "username: user, password: pass")
  if code != HTTP_OK {
    t.Fatalf(`Requested POST for '/login' with payload failed with %d`, code)
  }
}

func testPostWithoutPayload(t *testing.T) {
  server, err := createServer()
  if err != nil {
    t.Fatalf(`Failed to create server.`)
  }
  server.addPath("/login", "POST")
  server.run()
  code := server.POST("/login")
  if code != HTTP_BAD_REQUEST {
    t.Fatalf(`Requested POST for '/login' with no data accepted with %d`, code)
  }
}

func run() {
	return
}

func createServer() (string, error) {
  server := Server
  return server, nil
}
