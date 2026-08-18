package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// main reads the arguments and prints one HTTP/0.9 response.
func main() {
	// Take 3 arguments:
	// 1st: host
	// 2nd: port
	// 3rd: request
	args := os.Args[1:]

	if len(args) < 3 {
		log.Print("error: missing argument")
		os.Exit(2)
	}

	// An HTTP/0.9 client requests one document at a time.
	bytesSent, response, err := sendRequest(args[0], args[1], args[2])
	if err != nil {
		log.Printf("failed to make request: %v", err)
		os.Exit(1)
	}

	log.Printf("[debug] sent %d bytes", bytesSent)
	fmt.Print(response)
}

// sendRequest fetches one document over HTTP/0.9.
func sendRequest(host, port, rawRequest string) (int, string, error) {
	var response []byte
	var bytesSent int

	// A request is "GET", a space, the document address, and CRLF.
	request := fmt.Sprintf("GET %s\r\n", rawRequest)
	address := net.JoinHostPort(host, port)

	// The client opens a TCP connection to the given host and port.
	conn, err := net.DialTimeout("tcp", address, 20*time.Second)
	if err != nil {
		return bytesSent, "", fmt.Errorf(
			"unable to establish a connection with %s: %w",
			address,
			err,
		)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("[warn] failed to close connection: %v", err)
		}
	}()

	bytesSent, err = fmt.Fprint(conn, request)
	if err != nil {
		return bytesSent, "", fmt.Errorf("failed to send request %s: %w", request, err)
	}

	// The response ends when the server closes the connection.
	response, err = io.ReadAll(conn)
	if err != nil {
		return bytesSent, "", fmt.Errorf("failed to read the response: %w", err)
	}

	return bytesSent, string(response), nil
}
