package main

import (
	"fmt"
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
	server, cleanup := CreateServer(host, port, templatesPath, paths)
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

	fmt.Println("About to write data...")
	_, err = conn.Write([]byte(rt))
	fmt.Println(" data.")

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

	fmt.Println("About to write data...")
	_, err = conn.Write([]byte(rt))
	fmt.Println(" data.")

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

	fmt.Println("About to write data...")
	_, err = conn.Write([]byte(rt))
	fmt.Println(" data.")

	if err != nil {
		t.Fatalf(`Failed to write data to server %s`, err)
		return
	}

	// TODO: Read data back
	buff := make([]byte, 32768)
	_, err = conn.Read(buff)

	resString := string(buff[:])
	if !strings.Contains(resString, "400") {
		t.Fatalf(`Request failed with code %s, expected 400`, resString)
	}
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
	if server.paths["/"][0].method != "GET" || server.paths["/"][0].value != "Hello, World!" {
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_OK)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_OK)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_OK)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_OK)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_OK)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_OK)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_OK)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_OK)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_OK)
	}
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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}

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
	if !strings.Contains(resString, fmt.Sprint(HTTP_BAD_REQUEST)) {
		t.Fatalf(`Second request failed with code %s, expected %d`, resString, HTTP_BAD_REQUEST)
	}
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
		if !strings.Contains(resString, fmt.Sprint(HTTP_OK)) {
			t.Fatalf(`First request failed with code %s, expected %d`, resString, HTTP_OK)
		}
	}
}
