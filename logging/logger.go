package logging

import (
	"log"
	"io"
)


func NewInitLogger(handler io.Writer) *log.Logger {
	info := log.New(handler, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	return info
}

func NewErrorLogger(handler io.Writer) *log.Logger {
	ErrorLogger := log.New(handler, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return ErrorLogger
}
