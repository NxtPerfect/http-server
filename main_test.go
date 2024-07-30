package main

import "testing"

func TestValidPath(t *testing.T) {
  server, err := createServer(4043)
  if err != nil {
    t.Fatalf(`Failed to create server.`)
  }
  server.listen()
  code, err := server.GET("/")
  if err != nil {
    t.Fatalf(`Failed to send get.`)
  }
  if code != HTTP_OK {
    t.Fatalf(`Requested GET for '/' failed with %d`, code)
  }
}

func TestUnauthorized(t *testing.T) {
  server, err := createServer(4043)
  if err != nil {
    t.Fatalf(`Failed to create server.`)
  }
  server.addPath("/admin", GET)
  server.changePermission("/admin", "Role: Admin")
  server.listen()
  code, err := server.GET("/admin")
  if err != nil {
    t.Fatalf(`Failed to send get.`)
  }
  if code != HTTP_UNAUTHORIZED {
    t.Fatalf(`Requested GET for '/admin' without permissions accepted with %d`, code)
  }
}

func TestPostWithPayload(t *testing.T) {
  server, err := createServer(4043)
  if err != nil {
    t.Fatalf(`Failed to create server.`)
  }
  server.addPath("/login", POST)
  server.listen()
  code, err := server.POST("/login", "username: user, password: pass")
  if err != nil {
    t.Fatalf(`Failed to send post.`)
  }
  if code != HTTP_OK {
    t.Fatalf(`Requested POST for '/login' with payload failed with %d`, code)
  }
}

func TestPostWithoutPayload(t *testing.T) {
  server, err := createServer(4043)
  if err != nil {
    t.Fatalf(`Failed to create server.`)
  }
  server.addPath("/login", POST)
  server.listen()
  code, err := server.POST("/login", "")
  if err != nil {
    t.Fatalf(`Failed to send post.`)
  }
  if code != HTTP_BAD_REQUEST {
    t.Fatalf(`Requested POST for '/login' with no data accepted with %d`, code)
  }
}
