package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	State int // 0 for "initialized", 1 for "done"
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		State: 0,
	}

	buf := make([]byte, bufferSize)
	readToIndex := 0

	for r.State != 1 {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		// Handle real errors
		if err != nil && err != io.EOF {
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
			if readToIndex == 0 || r.State == 1 {
				break
			}

			return nil, errors.New("error: incomplete request")
		}
	}

	return r, nil
}

// TODO: Rewrite this function to return the number of bytes it consumed (and for functionality)
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

func (r *Request) parse(data []byte) (int, error) {
	// "Initialized" state
	if r.State == 0 {
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if numBytes == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = 1 // "Done" state
		return numBytes, nil
	} else if r.State == 1 {
		return 0, errors.New("error: trying to read data in a done state")
	} else {
		return 0, errors.New("error: unknown state")
	}
}