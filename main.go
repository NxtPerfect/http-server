package main

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

func (server *Server) listen() {
  panic("Not implemented yet")
	return
}

func createServer(port int) (Server, error) {
  var server Server
  server.port = port
  return server, nil
}

func (server *Server) addPath(path string, method int) (error) {
  // server.paths = { path: path, method: method }
  panic("Not implemented yet")
  return nil
}

func (server *Server) GET(path string) (int, error) {
  panic("Not implemented yet")
  return HTTP_OK, nil
}

func (server *Server) POST(path string, payload string) (int, error) {
  panic("Not implemented yet")
  return HTTP_OK, nil
}

func (server *Server) changePermission(path string, rule string) (error){
  panic("Not implemented yet")
  return nil
}
