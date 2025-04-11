package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type Writer struct {
	writerState writerState
	writer io.Writer
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

func NewWriter(w io.Writer) *Writer {
    return &Writer{
        writerState: writerStateStatusLine,
		writer: w,
    }
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("incorrect order for writing status")
	}
	defer func() {w.writerState = writerStateHeaders }()
	_, err := w.writer.Write(getStatusLine(statusCode))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("incorrect order for writing headers")
	}
	defer func() { w.writerState = writerStateBody }()
	for key, value := range headers {
		message := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.writer.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("error writing headers: %s", err.Error())
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("incorrect order for writing body")
	}
	return w.writer.Write(p)
}

// TODO: Write the below functions, concatenate the hex part with the written data
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
    if w.writerState != writerStateBody {
        return 0, fmt.Errorf("incorrect order for writing body")
    }
    hex := fmt.Sprintf("%x", len(p))
    w.writer.Write([]byte(hex + "\r\n"))
    n, err := w.writer.Write(p)
    if err != nil {
        return n, err
    }
    _, err = w.writer.Write([]byte("\r\n"))
    return n, err
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.writer.Write([]byte("0\r\n\r\n"))
}