package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte
	state requestState // 0 for "initialized", 1 for "done"
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

const crlf = "\r\n"
const bufferSize = 8

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		state: requestStateInitialized,
		Headers: make(map[string]string),
		Body: make([]byte, 0),
	}

	buf := make([]byte, bufferSize)
	readToIndex := 0

	for r.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		// Handle real errors
		if err != nil {
			if errors.Is(err, io.EOF) {
				if r.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", r.state, n)
				}
				break
			}
			return nil, err
		}
		readToIndex += n

		parsedBytes, parseErr := r.parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		if parsedBytes > 0 {
			copy(buf, buf[parsedBytes:readToIndex])
			readToIndex -= parsedBytes
		}

		if err == io.EOF {
			if readToIndex == 0 || r.state == requestStateDone {
				break
			}

			return nil, errors.New("error: incomplete request")
		}
	}

	return r, nil
}

func parseRequestLine(data []byte) (requestline *RequestLine, numBytes int, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]

	// Checks for only capital letters
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method: method,
		RequestTarget: requestTarget,
		HttpVersion: versionParts[1],
	}, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}

		if n == 0 {
			// need more data
			return 0, nil
		}

		r.RequestLine = *requestLine

		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLenStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil
		}
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return 0, fmt.Errorf("error: Content-Length could not be converted to an integer: %s", contentLenStr)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLen {
			return 0, fmt.Errorf("error: Content-Length too large")
		}
		if r.bodyLengthRead == contentLen {
			r.state = requestStateDone
		}
		return len(data), nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}

		// If no bytes were parsed, we need more data
		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}