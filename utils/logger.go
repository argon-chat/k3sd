// Package utils provides a Logger type and related utilities for logging cluster operations.
package utils

import (
	"fmt"
	"log"
	"time"
)

type Logger struct {
	// Stdout is the channel for standard output log messages.
	Stdout chan string
	// Stderr is the channel for error log messages.
	Stderr chan string
	// File is the channel for file log messages.
	File chan FileWithInfo
	// Cmd is the channel for command log messages.
	Cmd chan string
	// Id is the logger/session ID, used to distinguish log streams.
	Id string
}

type FileWithInfo struct {
	// FileName is the name of the file being logged.
	FileName string
	// Content is the content of the file being logged.
	Content string
}

// NewLogger creates and returns a new Logger instance with the given ID.
// The logger provides separate channels for stdout, stderr, file, and command logs.
//
// Parameters:
//
//	id: a string identifier for the logger/session (useful for grouping logs).
//
// Returns:
//
//	*Logger: a pointer to the new Logger instance.
func NewLogger(id string) *Logger {
	return &Logger{
		Stdout: make(chan string, 100),
		Stderr: make(chan string, 100),
		File:   make(chan FileWithInfo, 100),
		Cmd:    make(chan string, 100),
		Id:     id,
	}
}

// Log sends a formatted message to the logger's Stdout channel.
//
// Parameters:
//
//	format: a format string (as in fmt.Sprintf)
//	args: arguments for the format string
func (l *Logger) Log(format string, args ...interface{}) {
	l.Stdout <- fmt.Sprintf(format, args...)
}

// LogErr sends a formatted error message to the logger's Stderr channel.
//
// Parameters:
//
//	format: a format string (as in fmt.Sprintf)
//	args: arguments for the format string
func (l *Logger) LogErr(format string, args ...interface{}) {
	l.Stderr <- fmt.Sprintf(format, args...)
}

// LogFile sends a file's content to the logger's File channel.
//
// Parameters:
//
//	filePath: the name of the file being logged
//	content: the content of the file
func (l *Logger) LogFile(filePath, content string) {
	l.File <- FileWithInfo{FileName: filePath, Content: content}
}

// LogCmd sends a formatted command string to the logger's Cmd channel.
//
// Parameters:
//
//	format: a format string (as in fmt.Sprintf)
//	args: arguments for the format string
func (l *Logger) LogCmd(format string, args ...interface{}) {
	l.Cmd <- fmt.Sprintf(format, args...)
}

// LogWorker processes and prints all messages from the Stdout channel.
// If Verbose is false, it simply drains the channel with a delay (for background logging).
// Otherwise, it prints each message to the standard logger with a [stdout] prefix.
func (l *Logger) LogWorker() {
	if !Verbose {
		for range l.Stdout {
			time.Sleep(100 * time.Millisecond)
		}
		return
	}
	for logMessage := range l.Stdout {
		log.Printf("[stdout] %s", logMessage)
	}
}

// LogWorkerErr processes and prints all messages from the Stderr channel.
// Each message is printed to the standard logger with a [stderr] prefix.
func (l *Logger) LogWorkerErr() {
	for logMessage := range l.Stderr {
		log.Printf("[stderr] %s", logMessage)
	}
}

// LogWorkerFile processes and prints all file log messages from the File channel.
// Each file's content is printed with delimiters and the file name for clarity.
func (l *Logger) LogWorkerFile() {
	delimiter := "----------------------------------------"
	for logMessage := range l.File {
		strings := []string{delimiter, logMessage.FileName, delimiter, logMessage.Content, delimiter, logMessage.FileName, delimiter}
		log.Println("[FILE]")
		for _, s := range strings {
			log.Println(s)
		}
	}
}

// LogWorkerCmd processes and prints all command log messages from the Cmd channel.
// Each message is printed to the standard logger with a [CMD] prefix.
func (l *Logger) LogWorkerCmd() {
	for logMessage := range l.Cmd {
		log.Printf("[CMD] %s", logMessage)
	}
}
