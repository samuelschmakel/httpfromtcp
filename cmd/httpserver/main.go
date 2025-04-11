package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerHTTPbin(w, req)
		return
	}
	handler200(w, req)
	return
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
			<head>
				<title>400 Bad Request</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
		`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
		`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
		`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
	return
}

func handlerHTTPbin(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	path = "https://httpbin.org" + path
	
	resp, err := http.Get(path)
	if err != nil {
		fmt.Printf("error accessing server: %v\n", err)
		handler500(w, req)
		return
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Set("Transfer-Encoding", "chunked")
	w.WriteHeaders(h)

	buffer := make([]byte, 32) // buffer size of 32
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			w.WriteChunkedBody(buffer[:n])
		}
		fmt.Printf("The length of data being read is: %d\n", n)

		if err != nil {
			// This includes err == io.EOF
			break
		}
	}
	w.WriteChunkedBodyDone()

	fmt.Println(h, buffer)

}