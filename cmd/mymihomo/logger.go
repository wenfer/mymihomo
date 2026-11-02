package main

import (
	"fmt"
	"os"
	"time"
)

func timePrefix() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func logInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("[%s] %s\n", timePrefix(), msg)
}

func logWarn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "[%s] 警告: %s\n", timePrefix(), msg)
}

func logError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "[%s] 错误: %s\n", timePrefix(), msg)
}
