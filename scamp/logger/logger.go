package logger

import (
  "github.com/gudtech/scamp-go/scamp"
)

type Logger struct {}

func NewLogger() (l *Logger) {
  l = new(Logger)
  return
}

func Printf(fmtStr string, args ...interface{}) {
  str := fmt.Sprintf(fmtStr, ...args)
  
}