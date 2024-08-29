package main

import (
	"fmt"
	"net"
	"os"
	"testing"
)

func TestDefaultServerCreation(t *testing.T) {
	_, err := createDefaultServer()
	if err != nil {
		t.Fatalf(`Failed to create default server. %s`, err)
	}
}

func TestCustomServerCreation(t *testing.T) {
	_, err := createServer("1337", "/templates", []Path{{"/", "GET", "index.html"}})
	if err != nil {
		t.Fatalf(`Failed to create custom server. %s`, err)
	}
}

func TestConnectingToServer(t *testing.T) {
	server, err := createDefaultServer()
	if err != nil {
		t.Fatalf(`Failed to create default server. %s`, err)
	}

	server.listen()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)

	fmt.Println("Connected to server")

	rt := fmt.Sprintf("GET %v HTTP/1.1\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	_, err = conn.Write([]byte(rt))
	if err != nil {
		t.Fatalf(`Failed to connect to server %s %s`, err, conn)
	}
	fmt.Println("Wrote request to server")
}

func TestValidPath(t *testing.T) {
	server, err := createServer("4043", "/templates", []Path{{"/", "GET", "index.html"}})
	if err != nil {
		t.Fatalf(`Failed to create server.`)
	}
	server.listen()
	response, err := server.GET("/")
	if response.code != HTTP_OK {
		t.Fatalf(`Requested GET for '/' failed with %d.`, response.code)
	}
	correctContent, err := os.ReadFile("./templates/index.html")
	if err != nil {
		t.Fatalf(`Couldn't find correct file.`)
	}
	correctHtml := string(correctContent)
	if response.content != correctHtml {
		t.Fatalf(`Requested data is empty`)
	}
}

func TestUnauthorized(t *testing.T) {
	server, err := createServer("4043", "/temapltes", []Path{{"/admin", "GET", "You're an admin"}})
	if err != nil {
		t.Fatalf(`Failed to create server.`)
	}
	server.changePermission("/admin", "Role: Admin")
	server.listen()
	response, err := server.GET("/admin")
	if err != nil {
		t.Fatalf(`Failed to send get.`)
	}
	if response.code != HTTP_UNAUTHORIZED {
		t.Fatalf(`Requested GET for '/admin' without permissions accepted with %d`, response.code)
	}
	// TODO: check content
}

func TestPostWithPayload(t *testing.T) {
	server, err := createServer("4043", "/templates", []Path{{"/login", "POST", "Post without value"}})
	if err != nil {
		t.Fatalf(`Failed to create server.`)
	}
	server.listen()
	response, err := server.POST("/login", "username: user, password: pass")
	if err != nil {
		t.Fatalf(`Failed to send post.`)
	}
	if response.code != HTTP_OK {
		t.Fatalf(`Requested POST for '/login' with payload failed with %d`, response.code)
	}
	// TODO: check content
}

func TestPostWithoutPayload(t *testing.T) {
	server, err := createServer("4043", "/template", []Path{{"/login", "POST", "Post with payload"}})
	if err != nil {
		t.Fatalf(`Failed to create server.`)
	}
	server.listen()
	response, err := server.POST("/login", "")
	if err != nil {
		t.Fatalf(`Failed to send post.`)
	}
	if response.code != HTTP_BAD_REQUEST {
		t.Fatalf(`Requested POST for '/login' with no data accepted with %d`, response.code)
	}
	// TODO: check content
}
