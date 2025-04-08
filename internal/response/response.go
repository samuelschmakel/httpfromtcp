package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type Writer struct {
	W io.Writer
	writerState responseState
}

type responseState int

const (
	responseStateWritingStatusLine responseState = iota
	responseStateWritingHeaders
	responseStateWritingBody
	responseStateDone
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(map[string]string)

	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func NewWriter(w io.Writer) *Writer {
    return &Writer{
        W: w,
        writerState: responseStateWritingStatusLine,
    }
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != responseStateWritingHeaders {
		return fmt.Errorf("incorrect order for writing headers")
	}
	for key, value := range headers {
		message := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.W.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("error writing headers: %s", err.Error())
		}
	}
	_, err := w.W.Write([]byte("\r\n"))
	w.writerState = responseStateWritingBody
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != responseStateWritingBody {
		return 0, fmt.Errorf("incorrect order for writing body")
	}
	n, err := w.W.Write(p)
	w.writerState = responseStateDone
	return n, err
}