package common

import (
	"log"
	"os"
	"sync"
	"io"
)

var (
	globalLogger *SimpleLogger
	once sync.Once
)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

type LogConfig struct {
	FilePath string
}

func NewSimpleLogger(config ...*LogConfig) *SimpleLogger {
	if len(config) == 0 {
		once.Do(func() {
			globalLogger = &SimpleLogger{
				logger: log.New(os.Stdout, "", log.LstdFlags),
			}
		})
		return globalLogger
	}
	var writer io.Writer = os.Stdout
	if config[0].FilePath != "" {
		file, err := os.OpenFile(config[0].FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file: %v, using stdout instead", err)
		} else {
			writer = file
		}
	}
	return &SimpleLogger{
		logger: log.New(writer, "", log.LstdFlags),
	}
}

type SimpleLogger struct {
	logger *log.Logger
}

func (l *SimpleLogger) Infof(format string, args ...interface{}) {
	l.logger.Printf("[INFO] " + format, args...)
}

func (l *SimpleLogger) Warnf(format string, args ...interface{}) {
	l.logger.Printf("[WARN] " + format, args...)
}

func (l *SimpleLogger) Errorf(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] " + format, args...)
}
