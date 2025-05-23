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
	writerStateTrailers
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
	defer func() {w.writerState = writerStateTrailers}()
	return w.writer.Write(p)
}

// TODO: Write the below functions, concatenate the hex part with the written data
func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
    if w.writerState != writerStateBody {
        return 0, fmt.Errorf("incorrect order for writing body")
    }
	chunkSize := len(p)

	nTotal := 0
    n, err := fmt.Fprintf(w.writer, "%x\r\n", chunkSize)
	if err != nil {
		return nTotal, err
	}
	nTotal += n

    n, err = w.writer.Write(p)
    if err != nil {
        return nTotal, err
    }
	nTotal += n

    n, err = w.writer.Write([]byte("\r\n"))
	if err != nil {
		return nTotal, err
	}
	nTotal += n
    return nTotal, err
}
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	defer func() {w.writerState = writerStateTrailers}()
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != writerStateTrailers {
		fmt.Println("returning error")
		return fmt.Errorf("writing trailers out of order: %v", w.writerState)
	}
	for key, value := range h {
		message := fmt.Sprintf("%s: %s\r\n", key, value)
		fmt.Printf("message is: %s", message)
		_, err := w.writer.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("error writing headers: %s", err.Error())
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	return err
}