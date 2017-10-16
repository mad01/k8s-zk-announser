package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// LogInit func to set current log context
func LogInit(debug bool) {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})

	// Only log the warning severity or above.
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}
