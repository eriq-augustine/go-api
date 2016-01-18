package goapi;

import (
   "fmt"
)

type Logger interface {
   Panic(msg string)
   Fatal(msg string)
   Error(msg string)
   ErrorE(msg string, err error)
   Warn(msg string)
   WarnE(msg string, err error)
   Debug(msg string)
}

// A simple logger to use by default.
type ConsoleLogger struct {}

func (log ConsoleLogger) Panic(msg string) {
   panic(msg);
}

func (log ConsoleLogger) Fatal(msg string) {
   fmt.Println("Fatal: " + msg);
}

func (log ConsoleLogger) Error(msg string) {
   fmt.Println("Error: " + msg);
}

func (log ConsoleLogger) ErrorE(msg string, err error) {
   fmt.Printf("Error: %s [%v]", msg, err);
}

func (log ConsoleLogger) Warn(msg string) {
   fmt.Println("Warn: " + msg);
}

func (log ConsoleLogger) WarnE(msg string, err error) {
   fmt.Printf("Warn: %s [%v]\n", msg, err);
}

func (log ConsoleLogger) Debug(msg string) {
   fmt.Println("Debug: " + msg);
}
