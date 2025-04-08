package response

import (
	"fmt"
)

type StatusCode int


const(
	StatusCodeSuccess StatusCode = 200
	StatusCodeBadRequest StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func getStatusLine(statusCode StatusCode) []byte {
	reasonPhrase := ""
	switch statusCode {
	case StatusCodeSuccess:
		reasonPhrase = "OK"
	case StatusCodeBadRequest:
		reasonPhrase = "Bad Request"
	case StatusCodeInternalServerError:
		reasonPhrase = "Internal Server Error"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase))
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != responseStateWritingStatusLine {
		return fmt.Errorf("incorrect order for writing status")
	}
	_, err := w.W.Write(getStatusLine(statusCode))
	w.writerState = responseStateWritingHeaders
	return err
}