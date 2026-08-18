# HTTP/0.9 in Go

This repository contains a small HTTP/0.9 client and server written in Go. It follows the [original HTTP implementation described by the W3C](https://www.w3.org/Protocols/HTTP/AsImplemented.html).

The code is an educational example. HTTP/0.9 does not have the features expected from a modern HTTP client or server.

## What it implements

- The client opens a TCP connection to a host and port.
- A request contains `GET`, one space, and a document address.
- The client ends generated requests with CRLF. The server also accepts LF without CR.
- The server ignores words after the document address.
- Responses contain raw HTML without a status line or headers.
- Closing the connection marks the end of a response.
- Errors are returned as human-readable HTML.
- Client-aborted transfers are not recorded as server errors.
- The server does not keep request state after disconnecting.

## Requirements

- Go 1.26.2
- Permission to listen on TCP port 80

## Run the server

Run the server from its directory because it loads the HTML files using relative paths:

```bash
cd server
go run .
```

The server listens on port 80 and exposes `/` and `/index.html`.

## Run the client

Open another terminal and enter the client directory:

```bash
cd client
```

Pass the document address as the third argument:

```bash
go run . localhost 80 /index.html
```

You can also pipe a raw request into the client, like `nc`:

```bash
echo "GET /index.html" | go run . localhost 80
```

In pipe mode, the client forwards stdin without changing it. The server returns the HTML document and closes the connection.
