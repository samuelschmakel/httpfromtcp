package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading request: %s", err.Error())
	}
	str := string(bs)
	newLine := "\r\n"

	index := strings.Index(str, newLine)
	if index == -1 {
		return nil, errors.New("invalid request format")
	}
	requestLineStr := str[:index]

	requestLineData := strings.Split(requestLineStr, " ")
	if len(requestLineData) != 3 {
		return nil, fmt.Errorf("invalid request format. Actual length of request data is: %d", len(requestLineData))
	}

	requestLine := RequestLine{
		HttpVersion: requestLineData[2],
		RequestTarget: requestLineData[1],
		Method: requestLineData[0],
	}

	if !isAllUpper(requestLine.Method) {
		return nil, errors.New("invalid method format")
	}

	if requestLine.HttpVersion != "HTTP/1.1" {
		return nil, errors.New("unexpected HTTP version")
	}

	requestLine.HttpVersion = "1.1"

	if !strings.Contains(requestLine.RequestTarget, "/") {
		return nil, errors.New("invalid request target")
	}

	req := Request{
		RequestLine: requestLine,
	}
	return &req, nil
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}