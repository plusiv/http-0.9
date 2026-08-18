package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// main reads a request from an argument or stdin.
func main() {
	// Take host and port, plus an optional document address.
	// 1st: host
	// 2nd: port
	// 3rd: document address (optional)
	args := os.Args[1:]

	if len(args) < 2 {
		log.Print("error: missing argument")
		os.Exit(2)
	}

	// With two arguments, forward stdin unchanged like nc.
	var request io.Reader = os.Stdin
	if len(args) >= 3 {
		// A request is "GET", a space, the document address, and CRLF.
		request = strings.NewReader(fmt.Sprintf("GET %s\r\n", args[2]))
	}

	// An HTTP/0.9 client requests one document at a time.
	bytesSent, response, err := sendRequest(args[0], args[1], request)
	if err != nil {
		log.Printf("failed to make request: %v", err)
		os.Exit(1)
	}

	log.Printf("[debug] sent %d bytes", bytesSent)
	fmt.Print(response)
}

// sendRequest fetches one document over HTTP/0.9.
func sendRequest(host, port string, request io.Reader) (int64, string, error) {
	var response []byte
	var bytesSent int64

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

	// Stream the request into the TCP connection.
	bytesSent, err = io.Copy(conn, request)
	if err != nil {
		return bytesSent, "", fmt.Errorf("failed to send request: %w", err)
	}

	// The response ends when the server closes the connection.
	response, err = io.ReadAll(conn)
	if err != nil {
		return bytesSent, "", fmt.Errorf("failed to read the response: %w", err)
	}

	return bytesSent, string(response), nil
}
