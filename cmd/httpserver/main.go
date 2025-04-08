package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	html := ""
	if req.RequestLine.RequestTarget == "/yourproblem" {
		html = `
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
	`
		bytes := []byte(html)
		w.WriteStatusLine(response.StatusCodeBadRequest)
		h := response.GetDefaultHeaders(len(bytes))
		h.Override("content-type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(bytes)

	} else if req.RequestLine.RequestTarget == "/myproblem" {
		html = `
		<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
	`
		bytes := []byte(html)
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		h := response.GetDefaultHeaders(len(bytes))
		h.Override("content-type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(bytes)
	} else {
		html = `
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`
		bytes := []byte(html)
		w.WriteStatusLine(response.StatusCodeSuccess)
		h := response.GetDefaultHeaders(len(bytes))
		h.Override("content-type", "text/html")
		w.WriteHeaders(h)
		w.WriteBody(bytes)
	}
}