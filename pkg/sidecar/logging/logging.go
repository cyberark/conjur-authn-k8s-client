package logging

import (
	"log"
	"os"
)

var InfoLogger = log.New(os.Stdout, "INFO: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger = log.New(os.Stderr, "ERROR: ", log.LUTC|log.Ldate|log.Ltime|log.Lshortfile)
