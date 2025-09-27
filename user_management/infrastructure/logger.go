package infrastructure

import (
	"log"
	"os"
)

var Log *log.Logger

func InitLogger() {
	// Open or create the log file
	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a new logger that writes to the file
	Log = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}
