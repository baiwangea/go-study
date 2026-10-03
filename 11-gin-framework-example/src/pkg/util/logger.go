package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *log.Logger

func InitLogger() {
	now := time.Now()
	logDir := filepath.Join("logs", now.Format("2006-01"))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	logFile := filepath.Join(logDir, now.Format("2006-01-02")+".log")

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	logger = log.New(lumberjackLogger, "", log.LstdFlags)
}

func Info(format string, v ...interface{}) {
	logger.Printf(fmt.Sprintf("[INFO] %s", format), v...)
}

func Error(format string, v ...interface{}) {
	logger.Printf(fmt.Sprintf("[ERROR] %s", format), v...)
}
