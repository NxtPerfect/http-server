package main

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"testing"
)

func TestDefaultServerCreation(t *testing.T) {
	server, cleanup := CreateDefaultServer()
	defer cleanup()

	if isValidDefaultServer(t, server) != true {
		t.Fatalf(`Failed to create default server`)
	}
}

func TestCustomServerCreation(t *testing.T) {
	host := "127.0.0.1"
	port := "1337"
	templatesPath := "/custom_templates"
	paths := []Path{{"/custom", "GET", "index.html"}}
	server, cleanup := CreateServer(host, port, templatesPath, paths, true)
	defer cleanup()

	if server.host != host {
		t.Fatalf(`Failed to create default server Host error.`)
	}
	if server.port != port {
		t.Fatalf(`Failed to create default server Port error.`)
	}
	if server.templatesPath != templatesPath {
		t.Fatalf(`Failed to create default server TemplatesPaths error.`)
	}
	if len(server.paths) != len(paths) || len(server.paths["/custom"]) != 1 {
		t.Fatalf(`Failed to create default server Paths error.`)
	}
	if server.paths["/custom"][0].method != "GET" || server.paths["/custom"][0].value != "index.html" {
		t.Fatalf(`Failed to create default server Methods error.`)
	}
	if server.readyChan == nil {
		t.Fatalf(`Failed to create default server ReadyChan error.`)
	}
	if server.shutdownChan == nil {
		t.Fatalf(`Failed to create default server ShutdownChan error.`)
	}
}

func TestConnectToDefaultServer(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
}

func TestConnectToDefaultServerAndSendValidGetRequest(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])

	isValidServerResponse(t, resString, HTTP_OK)
}

func TestReceivingValidHTTPresponse(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])

	isValidServerResponse(t, resString, HTTP_OK)
}

func TestConnectToDefaultServerAndSendBadGetRequest(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt := fmt.Sprintf("DON'T %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)
}

func prepareAndRunDefaultServer(t *testing.T) func() {
	server, cleanup := CreateDefaultServer()

	if isValidDefaultServer(t, server) != true {
		t.Fatal("Failed to create default server.")
	}

	go func() {
		if err := server.Listen(); err != nil {
			t.Errorf("Server listen error: %v", err)
		}
	}()
	<-server.readyChan

	return cleanup
}
func isValidDefaultServer(t *testing.T, server Server) bool {
	if server.host != "127.0.0.1" {
		t.Errorf("Expected host to be 127.0.0.1, got %s", server.host)
		return false
	}
	if server.port != "1337" {
		t.Errorf("Expected port to be 1337, got %s", server.port)
		return false
	}
	if server.templatesPath != "/templates" {
		t.Errorf("Expected templatesPath to be /templates, got %s", server.templatesPath)
		return false
	}
	if len(server.paths) != 1 || len(server.paths["/"]) != 1 {
		t.Errorf("Expected one path for '/', got %d", len(server.paths))
		return false
	}
	if server.paths["/"][0].method != "GET" || server.paths["/"][0].value != "index.html" {
		t.Errorf("Unexpected path configuration for '/'")
		return false
	}
	if server.readyChan == nil {
		t.Errorf("readyChan is nil")
		return false
	}
	if server.shutdownChan == nil {
		t.Errorf("shutdownChan is nil")
		return false
	}
	return true
}

func isValidBenchmarkServer(t *testing.T, server Server) bool {
	if server.host != "127.0.0.1" {
		t.Errorf("Expected host to be 127.0.0.1, got %s", server.host)
		return false
	}
	if server.port != "1337" {
		t.Errorf("Expected port to be 1337, got %s", server.port)
		return false
	}
	if server.templatesPath != "/templates" {
		t.Errorf("Expected templatesPath to be /templates, got %s", server.templatesPath)
		return false
	}
	if len(server.paths) != 2 || len(server.paths["/"]) != 1 {
		t.Errorf("Expected one path for '/', got %d", len(server.paths))
		return false
	}
	if server.paths["/"][0].method != "GET" || server.paths["/"][0].value != "index.html" {
		t.Errorf("Unexpected path configuration for '/'")
		return false
	}
	if server.readyChan == nil {
		t.Errorf("readyChan is nil")
		return false
	}
	if server.shutdownChan == nil {
		t.Errorf("shutdownChan is nil")
		return false
	}
	return true
}

func TestTwoValidGetGetRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}

func TestTwoValidGetPostRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("POST %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}

func TestTwoValidPostPostRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("POST %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("POST %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}

func TestValidGetInvalidGetRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(500 * time.Millisecond))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)
}

func TestInvalidGetGetRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("GET %v HTTP/9.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)
}

func TestInvalidGetValidGetRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HP/1.0\r\n", "/")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}

func TestInvalidPostValidGetRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("POST %v HP/1.0\r\n", "/")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}

func TestInvalidPostPostRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("POST %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)

	// Second Request

	conn, err = net.Dial("tcp", serverAddressAndPort)
	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}

	rt = fmt.Sprintf("POST %v HTTP/9.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write second set of data to server %s`, err)
		return
	}

	_, err = conn.Read(buff)

	resString = string(buff[:])
	isValidServerResponse(t, resString, HTTP_BAD_REQUEST)
}

func TestThousandValidRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	for range 1000 {
		conn, err := net.Dial("tcp", serverAddressAndPort)
		defer func() {
			if conn != nil {
				fmt.Println("Closing down the connection client-side")
				conn.Close()
			}
		}()

		if err != nil {
			t.Fatalf(`Failed to connect to server %s`, err)
			return
		}
		rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
		rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
		rt += fmt.Sprintf("Connection: close\r\n")
		rt += fmt.Sprintf("\r\n")

		conn.SetDeadline(time.Now().Add(5 * time.Second))

		_, err = conn.Write([]byte(rt))

		if err != nil {
			t.Fatalf(`Failed to write first set of data to server %s`, err)
			return
		}

		buff := make([]byte, 32768)
		_, err = conn.Read(buff)

		resString := string(buff[:])
		isValidServerResponse(t, resString, HTTP_OK)
	}
}

func TestThousandInvalidRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	for range 1000 {
		conn, err := net.Dial("tcp", serverAddressAndPort)
		defer func() {
			if conn != nil {
				fmt.Println("Closing down the connection client-side")
				conn.Close()
			}
		}()

		if err != nil {
			t.Fatalf(`Failed to connect to server %s`, err)
			return
		}
		rt := fmt.Sprintf("GEEET %v HTP/9.0\r\n", "/")
		rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
		rt += fmt.Sprintf("Connection: close\r\n")
		rt += fmt.Sprintf("\r\n")

		conn.SetDeadline(time.Now().Add(5 * time.Second))

		_, err = conn.Write([]byte(rt))

		if err != nil {
			t.Fatalf(`Failed to write first set of data to server %s`, err)
			return
		}

		buff := make([]byte, 32768)
		_, err = conn.Read(buff)

		resString := string(buff[:])
		isValidServerResponse(t, resString, HTTP_BAD_REQUEST)
	}
}

func TestIfValidResponse(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
	rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_OK)
}
func isValidServerResponse(t *testing.T, response string, expectedCode int) {
	expectedStatus := getStatusForCode(expectedCode)
	firstResponseLine := fmt.Sprintf("HTTP/1.1 %v %s", expectedCode, expectedStatus)
	if !strings.Contains(response, firstResponseLine) {
		t.Fatalf(`Error in first line got: %s expected: %s`, response, firstResponseLine)
	}

	if expectedCode == HTTP_BAD_REQUEST || expectedCode == HTTP_NOT_FOUND {
		return
	}
	thirdResponseLine := "Server: Custom/Server"
	if !strings.Contains(response, thirdResponseLine) {
		t.Fatalf(`Error in third line got: %s expected: %s`, response, thirdResponseLine)
	}
	fifthResponseLine := "Content-Type: text/html"
	if !strings.Contains(response, fifthResponseLine) {
		t.Fatalf(`Error in fifth line got: %s expected: %s`, response, fifthResponseLine)
	}
	sixthResponseLine := "Content-Length:"
	if !strings.Contains(response, sixthResponseLine) {
		t.Fatalf(`Error in sixth line got: %s expected: %s`, response, sixthResponseLine)
	}
}

func getStatusForCode(code int) string {
	if code == 200 {
		return "OK"
	}
	if code == 404 {
		return "NOT FOUND"
	}
	if code == 400 {
		return "BAD REQUEST"
	}
	return "SERVER ERROR"
}

func TestInvalidPathRequests(t *testing.T) {
	cleanup := prepareAndRunDefaultServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	conn, err := net.Dial("tcp", serverAddressAndPort)
	defer func() {
		if conn != nil {
			fmt.Println("Closing down the connection client-side")
			conn.Close()
		}
	}()

	if err != nil {
		t.Fatalf(`Failed to connect to server %s`, err)
		return
	}
	rt := fmt.Sprintf("GET %v HTTP/1.0\r\n", "/fake/path/to/non/existing")
	rt += fmt.Sprintf("Connection: close\r\n")
	rt += fmt.Sprintf("\r\n")

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	_, err = conn.Write([]byte(rt))

	if err != nil {
		t.Fatalf(`Failed to write first set of data to server %s`, err)
		return
	}

	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	isValidServerResponse(t, resString, HTTP_NOT_FOUND)
}

func TestBenchmarkOneMillionRequests(t *testing.T) {
	cleanup := prepareAndRunBenchmarkServer(t)
	defer cleanup()

	serverAddressAndPort := "127.0.0.1:1337"

	fmt.Println("Starting one million benchmark...")

	start := time.Now()
	for range 1_000_000 {
		conn, err := net.Dial("tcp", serverAddressAndPort)
		defer func() {
			if conn != nil {
				fmt.Println("Closing down the connection client-side")
				conn.Close()
			}
		}()

		if err != nil {
			t.Fatalf(`Failed to connect to server %s`, err)
			return
		}

		chosenRequestType := rand.Intn(2)

		rt := ""
		if chosenRequestType == 0 {
			rt += fmt.Sprintf("GET %v HTTP/1.0\r\n", "/")
		} else {
			rt += fmt.Sprintf("POST %v HTTP/1.0\r\n", "/post")
		}
		rt += fmt.Sprintf("Host: %v\r\n", serverAddressAndPort)
		rt += fmt.Sprintf("Connection: close\r\n")
		rt += fmt.Sprintf("\r\n")

		conn.SetDeadline(time.Now().Add(5 * time.Second))

		_, err = conn.Write([]byte(rt))

		if err != nil {
			t.Fatalf(`Failed to write first set of data to server %s`, err)
			return
		}

		buff := make([]byte, 32768)
		_, err = conn.Read(buff)

		resString := string(buff[:])
		isValidServerResponse(t, resString, HTTP_OK)
		conn.Close()
	}

	elapsed := time.Since(start)
	fmt.Printf("One Million Requests Took: %s", elapsed)
}

func prepareAndRunBenchmarkServer(t *testing.T) func() {
	host := "127.0.0.1"
	port := "1337"
	templatesPath := "/templates"
	paths := []Path{{"/", "GET", "index.html"}, {"/post", "POST", "form.html"}}
	server, cleanup := CreateServer(host, port, templatesPath, paths, false)

	if isValidBenchmarkServer(t, server) != true {
		t.Fatal("Failed to create benchmark server.")
	}

	go func() {
		if err := server.Listen(); err != nil {
			t.Errorf("Server listen error: %v", err)
		}
	}()
	<-server.readyChan

	return cleanup
}
