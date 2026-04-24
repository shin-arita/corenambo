package logger

import "log"

func Info(format string, v ...any) {
	log.Printf("level=INFO "+format, v...)
}

func Warn(format string, v ...any) {
	log.Printf("level=WARN "+format, v...)
}

func Error(format string, v ...any) {
	log.Printf("level=ERROR "+format, v...)
}
