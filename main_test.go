package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"testing"
)

func TestDefaultServerCreation(t *testing.T) {
	_, err := CreateDefaultServer()
	if err != nil {
		t.Fatalf(`Failed to create default server. %s`, err)
	}
}

func TestCustomServerCreation(t *testing.T) {
	_, err := CreateServer("1337", "/templates", []Path{{"/", "GET", "index.html"}})
	if err != nil {
		t.Fatalf(`Failed to create custom server. %s`, err)
	}
}

func TestConnectToDefaultServer(t *testing.T) {
	prepareAndRunDefaultServer(t)
	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer conn.Close()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
}

func TestConnectToDefaultServerAndSendValidGetRequest(t *testing.T) {
	prepareAndRunDefaultServer(t)
	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer conn.Close()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.1\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	fmt.Println("About to write data...")
	_, err = conn.Write([]byte(rt))
	fmt.Println("Wrote data.")

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)
	fmt.Printf("Response: %s\n", buff)

	resString := string(buff[:])

	if !strings.Contains(resString, "200") {
		t.Fatalf(`Request failed with code %s, expected 200`, resString)
	}

}

func TestConnectToDefaultServerAndSendBadGetRequest(t *testing.T) {
	prepareAndRunDefaultServer(t)
	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer conn.Close()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("DON'T %v HTTP/1.1\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	fmt.Println("About to write data...")
	_, err = conn.Write([]byte(rt))
	fmt.Println("Wrote data.")

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}
	// TODO: Read data back
	buff := make([]byte, 32768)
	_, err = conn.Read(buff)
	fmt.Printf("Response: %s\n", buff)
	resString := string(buff[:])
	if !strings.Contains(resString, "400") {
		t.Fatalf(`Request failed with code %s, expected 400`, resString)
	}
}

func prepareAndRunDefaultServer(t *testing.T) {
	server, err := CreateDefaultServer()
	readyChan := make(chan struct{})
	server.readyChan = readyChan

	if err != nil {
		t.Fatalf(`Failed to create default server. %s`, err)
	}

	go server.Listen()
	<-readyChan
}

// func TestValidPath(t *testing.T) {
// 	server, err := CreateServer("4043", "/templates", []Path{{"/", "GET", "index.html"}})
// 	if err != nil {
// 		t.Fatalf(`Failed to create server.`)
// 	}
// 	server.Listen()
// 	response, err := server.GET("/")
// 	if response.code != HTTP_OK {
// 		t.Fatalf(`Requested GET for '/' failed with %d.`, response.code)
// 	}
// 	correctContent, err := os.ReadFile("./templates/index.html")
// 	if err != nil {
// 		t.Fatalf(`Couldn't find correct file.`)
// 	}
// 	correctHtml := string(correctContent)
// 	if response.content != correctHtml {
// 		t.Fatalf(`Requested data is empty`)
// 	}
// }
//
// func TestUnauthorized(t *testing.T) {
// 	server, err := CreateServer("4043", "/temapltes", []Path{{"/admin", "GET", "You're an admin"}})
// 	if err != nil {
// 		t.Fatalf(`Failed to create server.`)
// 	}
// 	server.ChangePermission("/admin", "Role: Admin")
// 	server.listen()
// 	response, err := server.GET("/admin")
// 	if err != nil {
// 		t.Fatalf(`Failed to send get.`)
// 	}
// 	if response.code != HTTP_UNAUTHORIZED {
// 		t.Fatalf(`Requested GET for '/admin' without permissions accepted with %d`, response.code)
// 	}
// 	// TODO: check content
// }
//
// func TestPostWithPayload(t *testing.T) {
// 	server, err := CreateServer("4043", "/templates", []Path{{"/login", "POST", "Post without value"}})
// 	if err != nil {
// 		t.Fatalf(`Failed to create server.`)
// 	}
// 	server.Listen()
// 	response, err := server.POST("/login", "username: user, password: pass")
// 	if err != nil {
// 		t.Fatalf(`Failed to send post.`)
// 	}
// 	if response.code != HTTP_OK {
// 		t.Fatalf(`Requested POST for '/login' with payload failed with %d`, response.code)
// 	}
// 	// TODO: check content
// }
//
// func TestPostWithoutPayload(t *testing.T) {
// 	server, err := CreateServer("4043", "/template", []Path{{"/login", "POST", "Post with payload"}})
// 	if err != nil {
// 		t.Fatalf(`Failed to create server.`)
// 	}
// 	server.Listen()
// 	response, err := server.POST("/login", "")
// 	if err != nil {
// 		t.Fatalf(`Failed to send post.`)
// 	}
// 	if response.code != HTTP_BAD_REQUEST {
// 		t.Fatalf(`Requested POST for '/login' with no data accepted with %d`, response.code)
// 	}
// 	// TODO: check content
// }
