package utils

import (
	"log"
	"os"
)

func NewLogger(prefix, file string) *log.Logger {
	logFile, err := os.Create(file)
	if err != nil {
		log.Fatal("Cannot create log file:", err)
	}
	return log.New(logFile, prefix, log.Ldate|log.Ltime|log.Lshortfile)
}
