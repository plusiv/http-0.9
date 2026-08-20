package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

const (
	portNumber        = 80
	errorFilePath     = "./error.html"
	inactivityTimeout = 15 * time.Second
)

var crlf []byte = []byte{'\r', '\n'}

// main starts the HTTP/0.9 server and accepts connections.
func main() {
	// HTTP uses port 80 when no other port is given.
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", portNumber))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", portNumber, err)
	}
	defer func() {
		if err := ln.Close(); err != nil {
			log.Printf("[err] failed to close listener: %v", err)
		}
	}()

	log.Print("[info] starting HTTP server")
	log.Print("[info] serving HTTP/0.9")

	// Map document addresses to HTML files.
	routes := map[string]string{
		"/":           "./index.html",
		"/index.html": "./index.html",
	}

	for {
		// The server accepts a TCP connection from the client.
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[err] failed to accept the connection: %v", err)
			continue
		}

		go handleConn(conn, routes)
	}
}

// handleConn reads one request and sends one response.
func handleConn(conn net.Conn, allowedRoutes map[string]string) {
	// The server closes the connection after sending the document.
	defer func() {
		if err := conn.Close(); err != nil && !isClientAbort(err) {
			log.Printf("[err] failed to close connection: %v", err)
		}
	}()

	if allowedRoutes == nil {
		log.Print("[warn] empty whitelist of routes")
	}

	// The server may close an inactive connection after about 15 seconds.
	if err := conn.SetReadDeadline(time.Now().Add(inactivityTimeout)); err != nil {
		log.Printf("[err] failed to set read deadline: %v", err)
		return
	}

	// A request ends with LF. The preceding CR is optional.
	reader := bufio.NewReader(conn)
	rawRequest, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Print("[warn] closing inactive connection")
			return
		}
		log.Printf("[err] unable to read request: %v", err)
		writeErrResponse(conn)
		return
	}

	method, route, err := parseRawReq(rawRequest)
	if err != nil {
		log.Printf("[err] unable to parse request: %v", err)
		// HTTP/0.9 sends errors as human-readable HTML.
		writeErrResponse(conn)
		return
	}

	if filePath, ok := allowedRoutes[route]; ok {
		log.Printf("[debug] handling request %s %s", method, route)

		// Tries to open the file.
		if file, err := loadFile(filePath); err == nil {
			writeResponse(conn, file)
			return
		} else {
			log.Printf("[err] unable to load file for response: %v", err)
		}
	} else {
		log.Printf("[err] invalid route: %s", route)
	}

	writeErrResponse(conn)
}

// parseRawReq extracts GET and the document address.
func parseRawReq(rawReq string) (string, string, error) {
	// A request starts with "GET" followed by one space.
	method, rawRoute, found := strings.Cut(rawReq, " ")
	if !found {
		return "", "", fmt.Errorf("malformed request")
	}

	if method != "GET" {
		return "", "", fmt.Errorf("unsupported method %s", method)
	}

	// The server ignores words after the document address.
	rawRoute, _, _ = strings.Cut(rawRoute, " ")

	if !strings.HasPrefix(rawRoute, "/") {
		return "", "", fmt.Errorf("invalid route %s", rawRoute)
	}

	route := strings.TrimSuffix(rawRoute, "\n")
	route = strings.TrimSuffix(route, "\r")

	return method, route, nil
}

// writeResponse sends an HTML document without headers.
func writeResponse(conn net.Conn, res []byte) {
	// Response lines end with LF. CR is optional.
	res = append(res, crlf...)

	// Stop waiting if the client does not read the response.
	if err := conn.SetWriteDeadline(time.Now().Add(inactivityTimeout)); err != nil {
		log.Printf("[err] failed to set write deadline: %v", err)
		return
	}

	// HTTP/0.9 sends the document as a raw byte stream.
	count, err := conn.Write(res)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Print("[warn] response timed out")
			return
		}
		// A client-aborted transfer is not recorded as an error.
		if !isClientAbort(err) {
			log.Printf("[err] failed to write response: %v", err)
		}
		return
	}

	log.Printf("[debug] wrote %d response bytes", count)
}

// writeErrResponse sends the HTML error document.
func writeErrResponse(conn net.Conn) {
	// HTTP/0.9 has no status code for error responses.
	errFile, err := loadFile(errorFilePath)
	if err == nil {
		writeResponse(conn, errFile)
	} else {
		log.Printf("[err] unable to write error response")
	}
}

// loadFile reads an HTML document from disk.
func loadFile(path string) ([]byte, error) {
	// The file becomes the response body sent to the client.
	return os.ReadFile(path)
}

// isClientAbort reports whether the client ended the transfer.
func isClientAbort(err error) bool {
	// The specification says not to record client aborts as errors.
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET)
}
