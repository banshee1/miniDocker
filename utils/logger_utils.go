package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

type loggerUtil struct {
}

var LoggerUtil = loggerUtil{}

// Init Init Logger Utils
func (util *loggerUtil) Init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
}

// Infof print format info message
func (util *loggerUtil) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Fatalf print format error message, then exit the process with code 1
func (util *loggerUtil) Fatalf(format string, args ...interface{}) {
	log.Errorf(format, args...)

	os.Exit(1)
}

// Errorf print format error message
func (util *loggerUtil) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args)
}
